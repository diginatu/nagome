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
A plugin should have corresponding directory and a plugin settings file named "plugin.yml" (describe later).

In the Nagome configure dir,

+   plugins
    +   plugin-name-1
        +   plugin.yml
        +   other files and directories you like
    +   ( plugin-name-2 )
        +   ( plugin.yml )

The plugin directory (plugin-name-1 above) name don't need to be same as its name.
But using the name connected with `-` is recommended.

### Main plugin

Nagome always creates one main plugin.  You can't make more.
Main plugin has plugin number 1.
Typically, it is used to a plugin that provides user interface, and the plugin executes Nagome.
So main plugin doesn't have plugin directory or fixed configuration file but you can pass -y(--ymlmain) command line option to specify the configuration file (same as plugin.yml in normal plugin).

You have to check below when you make a UI plugin.

+   Depend on the Domain "nagome_ui" and use all events as you can.
+   Do not use --dbgtostd command line option when you distribute it so users can see the log file later.

Differences between normal and main plugins:

+   Nagome will quit if the connection of main plugin is closed
+   Typically, main plugin executes Nagome.  Normal plugins are executed by Nagome.

plugin.yml
----------

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
    +   {{no}} : Plugin number (necessary in TCP).
    +   {{port}} : TCP port to connect (necessary in TCP).

+   nagomever : String.  Supporting version of Nagome (No effect).
+   depends : Array of string.  Domain of message that the plugin will receive (see Nagome message for more detail)

Connection
----------

Plugins communicate with Nagome process using JSON in stdin/out or TCP connection.
Each JSON message is sent line by line from Nagome.
Plugins don't have to keep this rule.

### stdin/out

To use stdin/out connection, set 'std' to 'method' in your plugin.yml.

You can directly use standard input and output for JSON communication.
Normal plugin can naturally use its stdin/out.
Main plugin should execute Nagome and grab the stdin/out of it.

### TCP

To use TCP connection, set 'tcp' to 'method' in your plugin.yml.

Different from stdin/out, TCP plugin should tell the plugin number to Nagome at first.
The number can be got as command line argument (see plugin.yml > exec).
Then, at the beginning of the connection, send a message like below.

~~~ json
{ 
    "domain": "nagome_direct",
    "command": "No",
    "content": {
        "no": YOUR_PLUGIN_NUM_HERE
    }
}
~~~

This message is special, so cannot be send at any time except this.


Nagome message
--------------

Nagome message is JSON message which is used in communication with Nagome.
All plugins and Nagome use this one message at each messaging.

The basic structure of a Nagome message is like blow.

+   Domain
+   Command
+   Content (optional)

### Domain

Basically, the message acts like pub-sub messaging.
A message sent from a plugin resend to other plugins which is domain plugin itself or depend on it.
"Depend" means that setting domain string in domain in the "plugin.yml".

#### Suffixed Domain

There is some special suffix.

##### @filter

The plugin that depends on a domain with this suffix can filter messages.
If there is a plugin that depends on filtering domain, a original message (without suffix) is added the suffix and sent to ONLY one plugin which depends filtering domain.

If the plugin wants to proceed the message, have to send the message with the suffix.
In this process, you can modify or just through, abort by not sending, also delay the message.

The suffixed message that passed all filtering plugins will broadcast to all plugins which depends the original domain.


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
