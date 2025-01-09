# UniqueID

This is an implemenation of the Snowflake unique id system described in "System Design Interview" by Alex Xu chapter 7. 


## Constraints

* IDs must be unique
* IDs are numerical values only
* IDs fit into 64-bits
* IDs are ordered by date.
* Ability to genreate over 10,000 unique IDs per second.

## Problems
It turned out to be much harder than I expected. If you store more than one atomic value, it is practically impossible to avoid duplicates. Also if you run out of values, what do you do? Originally I tried just incrementing the millisecond time, but that caused issues.

## Testing

I would run the test `TestUniqueID_NoDuplicates` as 1000 times:

```
go test -count=1000
```

This would be created 10 million unique ids in roughly 3.5 seconds. But most of the strategies that I used yielded duplicate values.

## Solution

In order to make absolutely sure that we get unique values, I do two things. First, there is only one value that is written and read atomically which is the proposed new unique id. In previous versions the epoch and inc where seperate and because there were two atomic writes, you'd occasionally get duplicates.

Second, if you exceed the maximum of 4096 ids per millisecond, then we sleep a millisecond. In the real world, this should _never_ happen because other operations are occurring, but I wanted this code to guarantee a unique id, and that was the only way.

I've run billions of unique ids with this code and haven't gotten a duplicate.
