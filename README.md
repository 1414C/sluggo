# sluggo

 - consider multicast like memcached?
https://gist.github.com/scottjbarr/255828
https://github.com/memcached/memcached/wiki


## model after zookeeper

Pc = process coordinator

- each new server assigns itself the highest id in the group (Pc + 1)
- new servers will have to read the id of the current coordinator from the db
- if the field is empty the server should write its own id (1) into the current coordinator db table

- at all times, the highest-id server petitions to be elected as the leader with the exception of joining servers

- each process monitors its next highest process neighbour by id

- timeout?
- check again
- timeout2?
    - remove successor and add old successor's successor as my successor updating topology map
    - was old successor the Pc?
        - yes - set flag and attempt to reach new successor (HB)  (recurse to top)
        - no  - clear flag and attempt to read new successor (HB) (recurse to top)

    - when a new successor is reachable
        - replacing the Pc     - send ELECTION MyPID (with updated topology map?)
        - not replacing the Pc - send REMOVE []OldSuccessorPID (with updated topology map?)


    - ELECTION - OriginPID = MyPID
    - Mnemonic == 'ELECTION'
        - send to sucessor
        - ...

