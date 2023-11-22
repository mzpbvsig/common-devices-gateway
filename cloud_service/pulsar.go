package cloud_service

import (
	"context"
	"reflect"
	"fmt"
	log "github.com/sirupsen/logrus"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/mzpbvsig/common-devices-gateway/bean"
)


type PulsarManager struct {
    PulsarClient pulsar.Client
	PulsarServer bean.PulsarServer
	RegisterProducer pulsar.Producer
	StateProducers map[string]pulsar.Producer
}

func NewPulsarManager(pulsarServer bean.PulsarServer) *PulsarManager {
    client := Connect(pulsarServer)

    return &PulsarManager{
        PulsarClient: client,
		PulsarServer: pulsarServer,
        StateProducers:    make(map[string]pulsar.Producer),
    }
}

func Connect(pulsarServer bean.PulsarServer) pulsar.Client {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:            pulsarServer.Url,
		Authentication: pulsar.NewAuthenticationToken(pulsarServer.Token),
	})

	if err != nil {
		log.Errorf("Connect pulsar error: %s", err)
		return nil
	}

	return client
}

func (pm *PulsarManager) CreateProducer(name string) (pulsar.Producer,error) {
	if pm.PulsarClient == nil {
        return nil, fmt.Errorf("Pulsar client is not initialized")
    }
	return pm.PulsarClient.CreateProducer(pulsar.ProducerOptions{
		Topic:                name,
		BatchingMaxSize:      1024 * 1024 * 5,
	})
}

func (pm *PulsarManager) Send(ctx context.Context, producer pulsar.Producer, msg []byte) bool {
	if producer == nil {
		log.Errorf("Producer is not initialized")
		return false
	}
	check := reflect.ValueOf(producer)
	if check.IsValid() {
		msgId, err := producer.Send(ctx, &pulsar.ProducerMessage{
			Payload: msg,
		})
		if err != nil {
			log.Errorf("Published message  to %s error: %s", producer.Topic(), err)
			return false
		} else {
			log.Printf("Published message %s to %s : %v", msg, producer.Topic(), msgId)
			return true
		}
	}
	log.Errorf("Producer is valid")
	return false
}

func (pm *PulsarManager) Register(ctx context.Context, msg []byte) bool {
	return pm.Send(ctx, pm.RegisterProducer, msg)
}

func (pm *PulsarManager) ReportState(ctx context.Context, gatewayId string, msg []byte) bool {
	return pm.Send(ctx, pm.StateProducers[gatewayId], msg)
}


func (pm *PulsarManager) ListenForEvents(handler func([]byte)) {
	go pm.createConsumers(handler)
}

func (pm *PulsarManager) CreateRegisterProducer(){
	registerTopicName :=  fmt.Sprintf("%s/register", pm.PulsarServer.Namespace)
    registerProducer, err := pm.CreateProducer(registerTopicName)
	if err != nil {
		log.Errorf("Create register producer error: %s", err)
		return 
	}
	pm.RegisterProducer = registerProducer

}

func (pm *PulsarManager) CreateStateProducers(deviceGateways []*bean.DeviceGateway){
	for _, deviceGateway := range deviceGateways {
		gatewayId := deviceGateway.Id

		if pm.StateProducers[gatewayId] != nil {
			continue
		}

		stateTopicName := fmt.Sprintf("%s/gateway-%s-state", pm.PulsarServer.Namespace, gatewayId)
		stateProducer, err := pm.CreateProducer(stateTopicName)
		if err != nil {
			log.Errorf("Create state producer error: %s", err)
		}
		pm.StateProducers[gatewayId] = stateProducer
	}
}

func (pm *PulsarManager) createConsumers(handler func([]byte)){
	if pm.PulsarClient == nil {
         log.Errorf("Pulsar client is not initialized")
		 return
    }

	topicName := pm.PulsarServer.Namespace + "/events" 
    consumer, err := pm.PulsarClient.Subscribe(pulsar.ConsumerOptions{
        Topics:           []string{topicName},
        SubscriptionName: "events",
		Type:             pulsar.Shared,
    })

    if err != nil {
        log.Errorf("Create Pulsar consumer for %s error: %s", topicName, err)
        return
    }

    defer consumer.Close()

    ctx := context.Background()

    for {
        msg, err := consumer.Receive(ctx)
        if err != nil {
            log.Errorf("Receive message from %s error: %s", topicName, err)
            continue
        }

        handler(msg.Payload())

        consumer.Ack(msg)
    }
}

