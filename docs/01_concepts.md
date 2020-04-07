# Concepts

## RandApp

RandApp is an upgrade to BaseApp that can store DKG-related messages and provides a querying interface for those messages. If you need to track messages sent by DKG participants (e.g. for accountability), you need to store those messages on-chain; this functionality is implemented by RandApp. Saved messages are kept with information about the DKG round that they belong to. 
