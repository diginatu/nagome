
Plugin
======

Nagome can have some plugins in the plugin folder, which is in the config directory.
Plugins communicate with Nagome process using JSON or MessagePack in std in/out or TCP connection.

Also, one application may create Nagome process and be dealt as a plugin.
This is mainly used as a client and provides user interfaces.

 + UI (also plugin)
   + Nagome
     + plugin1
     + plugin2


Message
-------

Message is sent by JSON or MessagePack.
There is two type of Message, event and query.

### Event

 + domain (Nagome)
 + func (CommentConnection)
 + type (got)
 + content (comment : "test")

### Query

 + command (CommentSend)
 + content (comment : "test")

