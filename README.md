# Livechat

Experimental chat application using UDP with live preview of the message your interlocutor enters. Powered by ![tcell](https://github.com/gdamore/tcell)

# Usage

```
go build
./livechat
```

# Idea
Text messaging is a common way to communicate. But unlike voice conversations, you have to wait for your interlocutor to finish typing before you can see the entire message. The livechat tries to remove difference between text and voice communication. You see a typing message before it would be sent.

# Todo
- Break NAT
- Use TURN-server for unbreakable NAT
