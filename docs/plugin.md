Plugin
======

Plugins can operate Nagome and catch and filter events trough the API.

Plugin types
------------

Nagome always creates one main plugin.
Other user defined plugin is called normal plugins.

+   main plugin
    +   Nagome
        +   normal plugin 1
        +   normal plugin 2

### Main plugin

Main plugin has plugin number 1.
Typically, it is used to a plugin that provides user interface.

Nagome will quit if the connection of main plugin is closed.
Main plugin executes Nagome process and passes settings via command line options.

#### Attention when you make a UI plugin

+   Depend on the Domain "nagome_ui" and use all events as you can.
+   Do not use --dbgtostd command line option when you distribute it so users can see the log later.

### Normal plugin

Normal plugins is placed in Plugins directory in Nagome configure directory.
A plugin should have corresponding directory and a plugin settings file named "plugin.yml" like below.

In the Nagome configure dir,

+   plugins
    +   plugin1
        +   plugin.yml
        +   other files and directories you like

#### plugin.yml

+   name : String
+   description : String
+   version : String
+   author : String
+   method : String.  "tcp" or "std".
+   exec : Array of string

    Nagome runs this code after loading this.
    Write your command line options as array elements.
    Following context will be replaced.

    +   {{path}} : Path to plugin directory.
    +   {{port}} : TCP port to connect (TCP only).
    +   {{no}} : Plugin number (Connection>TCP for detail).

+   nagomever : String.  Supporting version of Nagome (Not implemented yet).
+   depends : Array of string

Dependencies (see Message)

~~~ yaml
name: example
description: ""
version: "1.0"
author: diginatu
method: tcp
exec:
- ruby
- '{{path}}/example.rb'
- '{{port}}'
- '{{no}}'
nagomever: ""
depends:
- nagome
~~~

Connection
----------

Plugins communicate with Nagome process using JSON or MessagePack in stdin/out or TCP connection.

### stdin/out

### TCP

Message
-------

Message is sent in JSON or MessagePack.

+   Domain
+   Command
+   Content

An message sent from a plugin resend to other plugins which is domain plugin itself or depend on it.

### Example in JSON

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
