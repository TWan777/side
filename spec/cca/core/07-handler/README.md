# CCA Handler

The CCA Handler is an interface exposed to CAA applications that leverage CAA to implement their specific business logic. Typically, a CAA application is implemented as a standard Cosmos SDK module.

## Technical Specification

### Interface to hander inbound transaction:

```ts
function registerInboundTranasction(channelId: string, appId: string, tx: byte[]) {

}
```   
```ts
function onInboundExecuted(request: IntentRequest) {
}
```  
```ts
function onInboundConfirmed(request: IntentRequest) {

}
```  
```ts
function onInboundFinalized(request: IntentRequest) {

}
```  
```ts
function onInboundExpired(request: IntentRequest) {

}
```

### Interface to handle outbound transaction:

```ts
function registerOutboundSigningRequest(channelId: string, appId: string, tx: byte[]) {

}
```  
```ts
function onOutboundSigned(request: SigningRequest) {

}
```  
```ts
function onOutboundBroadcasted(request: SigningRequest) {

}
```  
```ts
function onOutboundExecuted(request: SigningRequest) {

}
```  
```ts
function onOutboundConfirmed(request: SigningRequest) {

}
```  
```ts
function onOutboundFinalized(request: SigningRequest) {

}
```  
