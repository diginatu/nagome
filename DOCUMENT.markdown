
Plugin
======

Nagome can have some plugins in the plugin folder, which is in the config directory.
Plugins communicate with Nagome process using JSON or MessagePack in stdin/out or TCP connection.

Also, one application may create Nagome process and be dealt as a plugin.
This is mainly used as a client and provides user interfaces.

 + UI (also plugin)
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
        + ngmplg.yml
        + other files

plugin.yml
----------

 + exec : Nagome runs this code at first when finish loading the plugin.
 + method : {"tcp", "std"}

~~~ yaml
---
name: "testplug"
description: "テストプラグイン"
depends:
  nagome: '1.0'
  plugin:
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

An message sent from plugin resend to other plugins which is domain plugin or depend on it.

### Example in JSON

Example of a comment message sent by nagome.

~~~ json
{
    "domain": "Nagome",
    "func": "CommentConnection",
    "type": "got",
    "content": {
        "date": "2016-04-10 14:11:39.823901 +0900 JST",
        "user": "ユーザ",
        "comment": "test",
        "iyayo": false
    }
}
~~~

Example of a sending comment message sent by a plugin.

~~~ json
{
    "domain": "Nagome",
    "func": "Query",
    "type": "CommentSend",
    "content": {
        "comment": "test",
        "iyayo": true
    }
}
~~~
