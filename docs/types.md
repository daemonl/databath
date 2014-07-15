Field Types
===========

CamelCases to check:
on_create
decimal_places
on_delete
on_update



Bool
----
TINYINT(1) NOT NULL
boolean

- type: bool
- default: (bool)

Date
----
DATE NULL
string "2014-12-31"

- type: date
- default: (string)
  - a concrete date(why...?)
  - #now
  - #start_month+-1d
  - #end_month+-1d


format for start and end
Start takes the first of whatever, end takes the last
Values can be day, week, month, year
then + adds another unit, which could be negative (+1d, +-1d), means add one day, subtract one day etc.
d: day
M: month
As per moment.js, i think.

DateTime
--------
INT(11) NULL
int UNIX timestamp

Special Values
- "#NOW" - current time

- type: datetime
- default: (int)


Enum
----
VARCHAR(x)

- type: enum

- length: 3
  - The length of the VARCHAR(x) definition, for the KEYS

- choices: {}
  - Map {string: string} of choices.

- default: (string)



File
----
VARCHAR(255) NULL
string

- type: file


Float
-----
FLOAT NULL
float

- type: float

- rhs: (string)
  - Text to display on the right of the field (e.g. "%", "hours")

- lhs: (string)
  - same as rhs, but on the left (e.g. "$")

- decimal_places: 2 (int)
  - the number of decimal places to autocorrect to and display with.
  - there is no '-1' for no rounding, floats will break`

ID
---
INT(11) UNSIGNED NOT NULL AUTO_INCREMENT
int

- type: id

Int
---
INT(11) UNSIGNED NULL
int

- type: int


Password
--------
VARCHAR(512) NULL (Salt + Hash)
string, does not return anything

- type: password

Ref
---
INT(11) UNSIGNED NULL


- type: ref

- collection: (string)
  - the name of the table / collection to link to.
  - links to the primary key (id) of the other collection

- on_delete: PREVENT (CASCADE|NULL|PREVENT)

- limit: {}
  - map of key and val to filter.
  - conditions are added to the remote collection.


Text
----
TEXT NULL

- type: text

String
------
VARCHAR(x)

- type: string

- length: 200 (int)
  - The length of the VARCHAR(x) definition
  - Validated in JavaScript as the max length.

- min_length: 0 (int)
  - When not 0, must be at least x characters

- checksum: (string)
  - The name of a checksum to run this field agains

- default: (string)

Timestamp
---------
TIMESTAMP DEFAULT CURRENT_TIMESTAMP
+? ON UPDATE CURRENT_TIMESTAMP
+? NULL / NOT NULL
string 2015-01-02 15:04:05


Special Values
- "#NOW" - current time

- type: timestamp

NOT YET IMPLEMENTED

KeyVal
------
TEXT NULL
json object

- type: key_val

RefId
-----
????
- type: ref_id

Address
-------
TEXT NULL
json object

- type: address
