Plugin
======

Plugins can operate Nagome and catch and filter events trough the API.

Plugin types
------------

Nagome always creates one main plugin.
Other user-defined plugin is called normal plugins.

+   main plugin
    +   Nagome
        +   normal plugin 1
        +   normal plugin 2

### Normal plugin

Normal plugins are placed in Plugins directory in Nagome configure directory.
A plugin should have corresponding directory and a plugin settings file named "plugin.yml" like below.

In the Nagome configure dir,

+   plugins
    +   plugin-name-1
        +   plugin.yml
        +   other files and directories you like
    +   ( plugin-name-2 )
        +   ( plugin.yml )

The plugin directory (plugin-name-1 above) is not need to be same as its name.
But using the name connected with `-` is recommended.

#### plugin.yml

Plugin configuration which is read at loading the plugin.

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

+   name : String
+   description : String
+   version : String
+   author : String
+   method : String.  "tcp" or "std".
+   exec : Array of string

    Nagome runs this code after loading this at startup.
    Write your command line options as array elements.
    Following context will be replaced.

    +   {{path}} : Path to plugin directory.
    +   {{no}} : Internal plugin number (necessary in TCP).
    +   {{port}} : TCP port to connect (necessary in TCP).

+   nagomever : String.  Supporting version of Nagome (Not implemented yet).
+   depends : Array of string.  Dependencies (see Nagome message for more detail)

### Main plugin

Nagome always creates one main plugin.
And you can't make more.
Main plugin has Internal plugin number 1.
Typically, it is used to a plugin that provides user interface.

Deferences between normal and main plugins:

+   Nagome will quit if the connection of main plugin is closed
+   Main plugin executes Nagome process and passes settings via command line options.

You have to check below when you make a UI plugin.

+   Depend on the Domain "nagome_ui" and use all events as you can.
+   Do not use --dbgtostd command line option when you distribute it so users can see the log file later.

Connection type
---------------

Plugins communicate with Nagome process using JSON in stdin/out or TCP connection.

### stdin/out

### TCP

Nagome message
--------------

Nagome message is JSON message which is used in communication with Nagome.
All plugins and Nagome use this one message at each messaging.

The basic structure of a Nagome message is like this.

+   Domain
+   Command
+   Content (optional)

### Domain

An message sent from a plugin resend to other plugins which is domain plugin itself or depend on it.

### Command

Command represents the message type in the Domain.

### Content

Content is structure for content of the Command.
So this may be unused in some Commands.

### Example

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
