# natstest
Scalability test for NATS message broker

# consumer deployment
cd consumer/
docker build --tag natstest-consumer .
docker image ls
docker tag <image> dockerhubuser/natstest-consumer
docker push dockerhubuser/natstest-consumer
kc apply -f natstest-consumer-statefulset.yaml

# producer
cd producer/
go build
natsServer=nats.example.com hubCount=200000 msgRate=10000 ./producer
2022/09/02 22:52:30 NATS producer test - hubCount 200000 natsServer nats.example.com msgRate 10000 msgBurst 100 sleepMs 10
2022/09/02 22:52:30 [Hub  0 ] Succeeded to connect to nats server
Sent messages to all  200000  hubs
Sent messages to all  200000  hubs
Sent messages to all  200000  hubs
