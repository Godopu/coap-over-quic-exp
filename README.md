# Mock CompressedCoAP for perforamance analysis

this project is currently in a technical preview state. It will be subject to a lot of modifications and is not yet for general purpose 

## Updating
- Now, We have started to update to move beyond testbed after performance analysis.
we will separate this directory as another repository soon. At that time, we will leave the url here

- we are changing protocol field to C-SEQ

## Todo
- [x] response message token handling
- [x] change token type
- [x] change token table name
- [x] make routine for recv message from SP
- [x] start to implement sp  
- [ ] make test code - doing
- [ ] make ack mechanism - doing 
- [ ] add configuration file

## Design
![](.README.assets/009a4295-87a0-49ef-afef-3ccfae4ba66c.png)



## Messaging
### Stream aggregation 
![](.README.assets/cb288c85-8288-4263-b89d-729c9860233e.png)

### Client to Server
![](.README.assets/583efb02-7f3a-4c48-ae5a-b417e4e28365.png)

### Server to Client
![](.README.assets/ec217357-0be8-4f58-ace9-54496f02ff5d.png)