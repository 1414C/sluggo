# sluggo

 - consider multicast like memcached?
https://gist.github.com/scottjbarr/255828
https://github.com/memcached/memcached/wiki


## model after zookeeper

Pc = process coordinator

- each new server assigns itself the highest id in the group (Pc + 1)
- new servers will have to read the id of the current coordinator from the db
- if the field is empty the server should write its own id (1) into the current coordinator db table

- at all times, the highest-id server petitions to be elected as the leader

- each process monitors its next highest process neighbour by id

