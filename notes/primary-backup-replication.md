# Primary Backup Replication
Backup and replication are very important subjects in keeping systems fault
tolerant in the wild. Many studies have been done and many schemes developed
that help craft fault tolerant systems that fail beautifully.

Replication schemes
- State transfer (primary/backup replication)
    in this method of replication, the primary sends a complete copy of the
    state of the primary at specific time intervals and sends that to the backup server.
    some of this things sent are contents of memory, cpu state, i/o device state
    but this things will result in enormous bandwidth usage to send the state
    everytime.
- replicated state transfer (state machine replication)
    this method models the servers as deterministic state machines that are kept
    in sync with each other by starting them at the same initial point. The
    assumption is that both servers recieve the same inputs in the same order.
    while this is rarely the case in realworld systems, the system must have 
    extra mechanisms for keeping servers in sync in the face of external inputs
    cause underterministic state in the system.

questions to answer on designing replication schemes
- what state to replicate
- how to keep primary/backup in sync
- how to handle cut-over and fail-over
- what to do in the face of anomalies
- how to handle new replicas

The paper in question today is the [VMware FT](https://pdos.csail.mit.edu/6.824/papers/vm-ft.pdf) paper.


