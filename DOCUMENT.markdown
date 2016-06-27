
Plugin
======

Nagome can have some plugins in the plugin folder, which is in the configuration directory.
Plugins communicate with Nagome process using JSON or MessagePack in stdin/out or TCP connection.

Also, one application may create Nagome process and be dealt as a plugin. (--standAlone option disable this)
This is mainly used as a client and provides user interfaces.

 + UI (UI plugin)
   + Nagome
     + plugin1
     + plugin2

Structure
---------

Non-UI plugins is placed in Plugins directory in Nagome configure directory.
UI plugin is not depend on the place.
Because it connects using stdin/out.

A plugin should have corresponding directory and a plugin management file named "plugin.yml."

 + [Nagome config]
   + plugins
     + plugin1
        + plugin.yml
        + other files

plugin.yml
----------

 + exec : Nagome runs this code at first when finish loading the plugin.
 + method : {"tcp", "std"}

~~~ yaml
---
name: "testplug"
description: "テストプラグイン"
nagomever: '1.0'
depends:
    - otherplug
version: '1.0'
author: "author"
exec: "ruby main.rb"
method: "tcp"
~~~

Message
-------

Message is sent in JSON or MessagePack.

 + domain
 + func
 + type
 + content

An message sent from a plugin resend to other plugins which is domain plugin itself or depend on it.

### Example in JSON

Example of a comment message sent by nagome.

~~~ json
{
    "Domain": "Nagome",
    "Func": "CommentConnection",
    "Command": "Got",
    "Content": {
        "Date": "2016-04-10 14:11:39.823901 +0900 JST",
        "User": "ユーザ",
        "Comment": "test",
        "Iyayo": false
    }
}
~~~

Example of a sending comment message sent by a plugin.

~~~ json
{
    "Domain": "Nagome",
    "Func": "Query",
    "Command": "CommentSend",
    "Content": {
        "Comment": "test",
        "Iyayo": true
    }
}
~~~
