# B3-go = Better Binary Buffers, for Go
B3 is a binary serializer which is easy like json, compact like msgpack, and powerful like protobuf,

B3 is a data serializer, it packs data structures to bytes & vice versa. It has:
* The schema power of protobuf, without the setup/compiler pain,
* The quick-start ease of json.dumps, but with support for datetimes,
* The compactness of msgpack, but without a large zoo of data types. 

With B3 you can fast-start with schema-less data (like json), and move to schemas (like protobuf) later & stay compatible. Or have ad-hoc json-like clients talk to rigorous protobuf-like servers without pain & suffering.

The small number of lovingly-handcrafted data types means often the only choice you need make is between Fast or Compact.

This is the Golang version. For more information & wire-format documentation, see the python reference implementation https://github.com/oddy/b3

This code is currently PRE-ALPHA, WIP/incomplete. 

* SCHED and DECIMAL support TBC
* Current composite support is dict-like to/from golang Structs. python/json dynamic-style list/dict with all-interface{}-s TBC.


