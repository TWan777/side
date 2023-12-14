# Baseapp

Telebase Apps should implement the `TeleBaseApp` interface, tailoring it to their specific business logic.

```ts
public interface TeleBaseApp {
   // on chain light client
   client LightClient;
   // route `IntentRequest` to handler
   function intentRequestRoute(intent: IntentRequest);
   // verify Inbound Tx and mint
   function onInboundAcknowledgement(txHash: string, proof: byte[]);
   // request TSS network for signing
   function requestSigning(signRequest: SigningRequest);
   // write TSS signature by TSS network leader
   function writeSignature(signRequest: SigningRequest, signature: string);
   // update status of tx execution
   function onOutboundAcknowledgement(txHash: string, proof: byte[]);       
}
```
