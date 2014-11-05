csgopugbot
==========

A Counter-Strike: Global Offensive IRC PUG bot written in Go. 

The PUG bot supports simultaneous PUG sessions, records in-game event statistics and has built-in web GUI for displaying PUG information. The bot runs without any game server related scripts and is configured via a JSON configuration file. A sample configuration file can be found in the project directory.

Please feel free to send through any feature requests, pull requests or issues as this project is being actively maintained.

Commands
==========

IRC commands are as follows;

- !pug [map] - Starts a PUG session on the desired map, if no map is specified, de_dust2 is selected by default.
- !join - Joins the user to the PUG session.
- !leave - Removes the user from the PUG session.
- !players - Lists the current users in the PUG.
- !stats - Currently not implemented.
- !say [message] - Sends a message to the CS server.

CS commands are issued by the PUG administrator and are as follows;

- !login [password] (required) - Authenticates the PUG administrator to issue further commands in-game.
- !map [map] - Changes map to desired map. NOTE: This can not be changed when the game has gone live.
- !request - Requests additional players from the IRC channel.
- !lo3 - Starts the PUG match.
- !cancelhalf - Cancels the current PUG half. The PUG administrator must type !lo3 to restart the half.
- !restart - Restarts the round. NOTE: This can not be issued when the game has gone live.
- !say [message] - Sends a message to the IRC server.
