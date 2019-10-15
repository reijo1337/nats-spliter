# nats-spliter
Simple nats message spliter

Splitter gets messages from sourse NATS streaming, looks at separator value, which name gets from `SEPARATOR_NAME` env value. Then service choose destinations NATS by this separator. 

Splitter send message to all destination NATS if `SEND_TO_ALL` is true (as by default) and:

1. Service gets error on parsing message
2. Message has no separator
3. There is no destination NATS for message separator