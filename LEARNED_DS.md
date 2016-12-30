# Distributed System

My server was made to run on a computer with multiple cores not multiple
computers which will pose a limit on the number users, The number 1 restriction is proberly
the number of ws connections.

Running a distributed system allow 3 important properties. 

* **Scalable** To be scalable the system must able to run on multible computers.

* **24x7** To run 24x7 every part must be replacable without downtime.

* **Recoverable** The loss of one machines most be recoverable without downtime.

The current system data:

* **Clients** a map of of clients primary use securing unique user names and restricting
login to one. We could allow for multiple login but we still need to connect to one identity.
The map persistent as it is saved when the server is closed.

* **Players** a map of players primary use to close down nicely. The players holds wss
connection, lists of invites, the game or watch games.

* **Tables** a map of tables primary use restrict the player to start one game.
the table hold a game (two players) and a list of player watching.

* **Public list** a map of players and tables comunication channels and id's

* **Saved games** a map of saved games to be resumed later.
The map persistent as it is saved when the server is closed.

The distributed systems properties pose difference but similar demands on the data.

## Scalable

The websocket connections must be able to run on multiple machines so we must
have a server which include a *Http server* and the Players and Tables and be able
to run muliple instances of it. The remaning data Clients, Public list and saved games
must be available to the http servers as a service by the *Data server*.

## 24x7 

To run 24x7 all parts of the system must be possible upgrade while running. Therefore
all parts must be able to run two version of the same program at the same time which 
means that all parts must be able to run at least two instances. For the scalable https
server it is not a problem but the Data server need new features.

* Synchronise to allign the two instances. 

* Handover transfer the clients in this case the https servers.

If we do not want to block the https servers we must queue the https data
while synchronise and then have a period where the new instance recieve queued data
from old instance and the http servers. The order of data from two different machines
can't be garantied so we need a way of ordering the incoming data. A time stamp of 
the http server data should be ennough to prevent a late buffered piece of http
server data to overwrite a new piece of http server data.

## Recoverable

Persistent data gives some protections from failure as a crash program can recover with 
the persistented data but it takes time. Persistent is data that exist outside the memory
and therefore exist after memory is lost. The memory is lost when the program stops or 
crashes, when machine(vm) stops or crashes. The persistent data is lost when the filesystem
is lost.Persistent data also makes it possible to hold more data than can be in memory.
When running 24x7 on vm's with big memories persistent data does loses some of its value.
Having the data mirrowed in memory on a different machine is better that having data
mirrowed on a file.

The current system can only recover from controled close down. The Data server does not
hold all then information we want to conserve if a Http server crashes. The games that
is being played must also be mirrowed so incase a Http server crashes the games is not
lost. A log of finished game may also be need. This adds two new services *Game backup*
and *Game log*. The Game backup must be scalable and does not need to be mirrowed as the
data already exist the system can run without for a few seconds. The Game log data need to
be mirrowed and the service dublicated in case of a crash. The Data server need to be
dublicated and mirrowed incase of a crash.

# System

Mirrowed data is a developer nightmare as it create a synchronisation problem and if 
perferct synchronisation is demanded the system must lock which prevent scale.
Therfore the system must have a way of dealing with inconsitency when they happen,
of course improving synchronisation reduce inconsitentcies.

When a service have to be dublicated and the data mirrowed to run 24x7 and provide fast
recover it make sense to use the service to improve the system capacity if possible.
If we know that some data is need more some places we can make the synchronisation
accordingly. If we know a user uses one data center we can have her data there in perfect
state and use another data center as backup with a delay. When the user switch data center
we may get some synchronisation problem but it should not happen often.

A destributed version of the gameserver would consist of at least two system that provide
failover support for each other. A system consist of one Data server and one Game log 
one or more Game backup and one or more Http server.
The Game backup is indepent of the other system if it crashes just start a new and the system
should survive without it. The Http server should know the other systems Data server and 
Game log in case of a fail. The Data servers must synchronise information but the Game log 
could save the games in individual files.

# Container manager system.

Container is only a easy consistent way of deploying programs to run a distributed
system you need more [kubernetes](http://kubernetes.io/docs/whatisk8s/) may be the 
way to go. TODO look in to kubernetes.

# Message System

To run on multiple machines we need a comunication frame work. 
the current system use go channels to comunicate and it work nicely. The system
use channels that blocks until sender and reciever is ready to exchange messages.
It use a selector to use between to ready channels and the sender can send a special
close channel signal that does not block the sender. Buffers can also be used in that
case the sender can send more messages before it have to wait and the messages is
queued on the reciever side.

The messages system between machines must be added in between the channels of the
machines how big a problem that would be time will tell. I think all messages systems
must suffer from the last messages uncertaincy. The fait of a messages that is send without
expecting a return (the last message) is always unknown.

## [ZeroMq](http://zeromq.org/)

