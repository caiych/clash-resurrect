# clash-resurrect

## What it does now

Check clash RESTful API and see if it's still responding, if it's not -- kill it. 
NOTE: This is depending on other serivce(e.g. supervisor) to bring it up again.

### Why?

The clash on my raspberry pi becomes frozen sometimes. 
So I started with a small script trying to reboot it when it happens.
Then the settings are lost. I'm writing this project to 

### Why not write a supervisor that watches clash?

I'm not confident enough to write a deamon that out perform(in terms of reliability) any standard daemon tools.

## TODO

* Proxy group(which is the most importtant) is persisted. Also persist log level, mode, etc.
* Unit test and other code quality improvments.
