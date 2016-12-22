Plugin
======

Plugins can operate Nagome and catch and filter events trough [Nagome Message](nagome_message.md).

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

### Main plugin

Nagome always creates one main plugin.  You can't make more.
Main plugin has plugin number 0.
Typically, it is used to a plugin that provides user interface, and the plugin executes Nagome.
So main plugin doesn't have plugin directory or fixed configuration file but you can pass -y(--ymlmain) command line option to specify the configuration file (same as plugin.yml in normal plugin).

You have to check below when you make a UI plugin.

+   Subscribe the Domain "nagome_ui" and use all events as you can.
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
subscribe:
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
+   subscribe : Array of string.  Domain of message that the plugin will receive (see Nagome message for more detail)

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

Example
-------

[StdExample](https://github.com/diginatu/nagome-stdexample)

Example of standard io plugin written in ruby/gtk3.
