Nagome message
==============

Nagome message is JSON message which is used in communication with Nagome.
All plugins and Nagome use this one message at each messaging.

The basic structure of a Nagome message is like blow.

+   Domain
+   Command
+   Content (optional)

All types of Nagome message are found in [api.go](../api.go).

Domain
------

Basically, the message acts like pub-sub messaging.
A message sent from a plugin resend to other plugins which is domain plugin itself or subscribe it.
"Subscribe" means that the domain name in the message is set to "subscribe" in the "plugin.yml".

### Suffixed Domain

There is some special suffix.

#### @filter

The plugin describing a domain with this suffix can filter messages.
If there is a plugin that describes on filtering domain, a original message (without suffix) is added the suffix and sent to ONLY one plugin which describes filtering domain.

If the plugin wants to proceed the message, have to send the message with the suffix.
In this process, you can modify or just through, abort by not sending, also delay the message.

The suffixed message that passed all filtering plugins will broadcast to all plugins which describes the original domain.


Command
-------

Command represents the message type in the Domain.

Content
-------

Content is structure for content of the Command.
So this may be unused in some Commands.

Example
-------

Example of a comment message sent by nagome.

~~~ json
{
    "domain":"nagome_comment",
    "command":"Comment.Got",
    "content":{
        "No":12,
        "Date":"2016-09-08T16:56:54.786312+09:00",
        "UserID":"1234",
        "UserName":"user",
        "Raw":"test",
        "Comment":"test",
        "IsPremium":false,
        "IsBroadcaster":false,
        "IsStaff":false,
        "IsAnonymity":false,
        "Score":0
    }
}
~~~

Example of a sending comment message sent by a plugin.

~~~ json
{
    "Domain": "nagome_query",
    "Command": "Broad.SendComment",
    "Content": {
        "Text": "test",
        "Iyayo": true
    }
}
~~~
