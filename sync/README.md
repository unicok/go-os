# Sync interface

The sync interface provides a way to coordinate across a number of nodes. This can be used 
for leadership election, membership consensus, etc. It's a building block for application 
synchronization.

We want the ability to choose between CP and AP where CP is useful for transactional and leader 
behaviour and AP for eventually consistent semantics.
