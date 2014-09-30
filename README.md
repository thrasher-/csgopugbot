csgopugbot
==========

A Counter-Strike: Global Offensive IRC PUGbot written in Go.

The bot reads from a config.ini file which provides it all the neccassary information to function, a sample config file can be located in the project directory.

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
