# Google Voice Client & Matrix Bridge

## Features

- [x] Login to Google Voice
- [x] Send SMS messages
- [x] Fetch inbox/threads
- [x] Basic Matrix bridge
- [x] Initial code to show how to receive SMS messages in real-time
- [ ] Double puppeting
- [ ] Storing data in local database
- [ ] Testing & documentation

## Usage

1. Setup the config.yaml file, currently configured for @jkeating:beeper.com's account
1. Start the bridge
   ```
   $ cd matrix_googlevoice
   $ go run main.go
   ```
1. Open a chat with the bot
   ```
   /query @gvoicebot:beeper.local
   ```
1. Interact with the bot
   ```
   help
   login ....
   send-message +15555551234 Hello World
   latest-message
   ```
   
