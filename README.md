# Livechat

Experimental chat application using UDP with live preview of the message your interlocutor enters. Powered by [tcell](https://github.com/gdamore/tcell)

## Usage

```
go build
./livechat
```
Select via arrows `Start chatting` and press [Enter]. After server will be created you see server address. Give this address to person with you want to chat. When you know recipient address, you can create chat. Select `New chat`, press [Enter], then input recipient address and press [Enter] again. Start typing message and you recipient will see new chat below his server address.

## Idea
Text messaging is a common way to communicate. But unlike voice conversations, you have to wait for your interlocutor to finish typing before you can see the entire message. The livechat tries to remove difference between text and voice communication. You see a typing message before it would be sent.

It uses Free WebRTC TURN Server powered by [Metered Video](https://www.metered.ca/).

## To do
* Scrollable chat history
* Add cryptography
