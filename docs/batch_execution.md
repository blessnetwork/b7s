## Batch Execution

Batch execution request allows specifying many execution requests in a single call.
Batch execution request specifies execution of a single Bless function with many different arguments lists.

## Batch Execution Flow

A Batch Execution Request consist of a template and a list of arguments for each individual execution.
Template resembles an ordinary execution request - it specifies the Bless function to be executed and the configuration.
Arguments are specified in a list of arguments lists.

If we have a Batch Execution Request which will produce 100 executions - we will have 1 template and a list with 100 arguments lists.

The input Batch Execution Request is termed a `Batch` in the head node.
Each combination of a template and an argument list is a `Work Item`.
Thus, a Batch which specifies 100 individual executions is a Batch consisting of 100 Work Items.

When a Batch Execution Request is received, the head node does a roll call.
The number of workers it waits for is the number specified in the input request.

After enough workers have responded, head node will split the work items into `Chunks`.
Number of chunks is equal to the number of workers that have been chosen for the work.

To recap - `Batch` consists of individual `Work Items`.
After workers have been chosen, we have a `Batch` split into `Chunks`, where each chunk has a worker it was assigned to, and `Chunk` consists of `Work Items`.

Currently Work Items are assigned to Chunks/workers in a round-robbin manner.
One Work Item is only assigned to a single Chunk.

Chunks are then sent to chosen workers for execution.

When a Batch is created, a handle - UUID - is returned to the user.
The user can use this handle to query for the current status/progress of the execution.

As workers are completing their chunks, they respond to the head node with their results.
Head node processes these messages and updates corresponding work items in the DB.

For each work item, we store the standard output of the execution, and the status of the execution - pass/fail.

When processing a result for a chunk, we do some validation:
- was the sending node indeed the node that was assigned this chunk
- was the work item whose result we're processing indeed part of the chunk in question
- was the work item really in the "in progress" state - to prevent overwriting already executed items

When we process a chunk result, we do additional checks and, if all of the items in a chunk have been processed, we mark the chunk as `DONE`.

### Head Node Boot Sequence

After a head node boots, it should restart any past executions.

When the head node main loop starts, it will allow for a short delay to allow connections to workers to be re-established.
This delay is one minute, by default.

After this, it will try to resume any batches that have not been completed.
First, we will query the Batch Store for any Batches in the state `CREATED` or `IN PROGRESS`.

Then, it will iterate through the list of batches and attempt to resume them.

Batch resume involves finding all work items belonging to the batch that are in the `CREATED` or `FAILED` state.
You can read more about this in the [Work Item States](#work-item-states) section.

For all `FAILED` work items, we check if they have passed the failure threshold, and, if so, we now mark them as `PERMANENTLY FAILED`.
For all other work items, we start another round of batch execution - which includes doing a roll call to find new set of workers.
We will create new chunks and assign these work items to them, and send them to workers for execution.

This process can repeat until all work items are either `DONE` or `PERMANENTLY FAILED`.
At this point, Batch is marked as complete.

This process is also done at regular intervals, controlled by the _Requeue Interval_ config option.
By default this interval is 1 hour.

One important thing to note is - batch is considered complete only when a resume is attempted.
This will happen either on node boot or on next requeue interval.
This is done so we do less frequent querying.

### Batch Execution Result Format

When returning Batch Execution result, we return a list (actually a map) of individual chunks comprising that batch.
Chunk IDs are v4 UUIDs and have no meaning.

For each Chunk, we include the ID of the peer that performed that execution and the list of results.
Results of Work items in a chunk are presented in a map, mapping work item ID to the result of the execution.

### Work Item ID

We must provide a way to map one specific instance of execution to its result.
For example we want to execute function found at CID `c`, specifically `f.wasm`, with two sets of arguments: `--input-arg1 a1 --input-arg2 a2` and `--input-arg2 b1 --input-arg2 b2`.
How do we find out which result belongs to which?

For this, we introduced a work item ID which can be derived from its inputs.
This way, by knowing the execution inputs, you can derive the ID that you need to lookup in the set of results.

Work Item ID is calculated as an md5 checksum of the function invocation, presented as `C/f.wasm <space-separated-argument-list>`.
For the first execution above, that would mean `c/f.wasm --input-arg1 a1 --input-arg2 a2`.

For a real world example, `bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm https://example.com/dir1/dir2/resource/some-random-slug-1` produces an MD5 hash of `268a4145a50ade48aed2b1147d3518c6`.
Thus, when iterating results a chunk, we only need to be on the lookout for this specific key to find our results.

## Stores

Batch Execution requests can consist of many individual Work items and their execution can take a long time.
To handle this head node persists this information to a `Batch Store`.

### MBS - Mongo Batch Store

Mongo Batch Store will use an external MongoDB server as a backing store.
This allows head node to save information to a MongoDB and resume work even after it has restarted.

The downside is that there needs to be an externally configured (and reachable) MongoDB server that the head node can use.

When a Mongo Batch Store is initialized, it will try to connect and validate connection to a MongoDB server (by issuing a Ping request).
If the connection is successful, it will proceed to create/configure the collections we will use.
By default Mongo Batch Store will try to create collections to use:

- `b7s-batches`
- `b7s-batch-chunks`
- `b7s-batch-work-items`

We have created schemas for these collections, to ensure proper validation.
Schema files can be found [here](stores/batch-store/mbs/validation).

If the collections already exist with the same configuration, the attempt to create the collections again is a no-op.
An attempt to create a collection with a different schema than the one that already exists will fail.
However, because of how MongoDB Go driver operates, identical schemas can be encoded differently - e.g. different order of fields.
This may lead to errors that are false positives.

As we assume some exclusivity on the MongoDB server, we do ignore these specific errors as we assume the cause for them is the one listed above.

To specify a MongoDB server we can use command-line arguments:

```console
$ ./node --batch-db-server mongodb://172.26.176.1:27017/ --batch-db-name b7s-db # other CLI arguments
```

Alternatively, MongoDB server can be specified using the standard head node config file:

```yaml
head:
  batch:
    server: mongodb://172.26.176.1:27017/
    db-name: b7s-db
```

### IBS - In-Memory Batch Store

MongoDB Batch Store relies on having an externally configured and reachable MongoDB server.
This is a hindrance to users wanting to run a head node without external dependencies.

In order to provide this, we include an In-Memory Batch Store.
This Batch Store keeps all data in working memory of the head node.

Obviously, this Batch Store does not survive head node restarts and all content is gone after the head node has stopped.

### Work Item Database ID

Function invocation does not have to be universally unique.
Many users can execute the same Bless function with the same arguments.

In order to avoid collisions, individual work items are stored in the DB using the `<batch-id>/<work-item-id>` format.

You can read more about the work item ID format [here](#work-item-id).

### Work Item States

Work Item can have a different number of states during its lifetime.

- `0` - `CREATED` - Work item was saved but not yet assigned to a chunk or execution has been started
- `1` - `IN PROGRESS` - Work Item was created and sent to a worker for execution
- `100` - `DONE` - Work item was executed by a Worker as part of a chunk. The result of the execution resulted in a successful exit code (0).
- `-1` - `FAILED` - Work item was executed by a Worker as part of a chunk. The result of the execution resulted in a non-zero exit code.
- `-2` - `PERMANENTLY FAILED` - Work Item was executed by a Worker or Workers multiple times, after which we will no longer retry execution.

Execution of a Work item can fail for various reasons.
For example, an HTTP call could fail because the IP of a Worker was added to a blocklist by the target server, or a resource was not available.
In that case it makes sense to retry that execution, potentially by a different Worker.
However, we don't want to retry execution indefinitely.

We allow the user to specify a `max_attempts` field, which is an unsigned number specifying how many times will we try to execute a single Work item.
After this threshold has been reached, we give up on the individual work item.

Note that a user can be willing to specify an arbitrarily large limit for Work items.
From the perspective of both head and worker nodes, this means many retries and potentially wasted resources.

To prevent this, we allow the head node operator to specify a global limit on number of retries.
When we consider if a work item should be retried, we use the  _lower_ of the two limits.

By default the Head node will give up after 10 failed attempts.


### Batch Execution REST API Call

Below is a curl invocation starting a batch execution request for a function `bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q`, which requires 4 nodes, and a list of four lists of CLI arguments for the Bless function.
Note that in this case each argument list consists of a single argument (a URL) but that does not have to be the case.

```sh
curl --location '127.0.0.1:8080/api/v1/functions/execute/batch' \
--header 'Content-Type: application/json' \
--data '{
    "template": {
        "function_id": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q",
        "method": "echo.wasm",
        "config": {
           "number_of_nodes": 4
        }
    },
    "arguments": [
        ["https://example.com/dir1/dir2/resource/some-random-slug-0"],
        ["https://example.com/dir1/dir2/resource/some-random-slug-1"],
        ["https://example.com/dir1/dir2/resource/some-random-slug-2"],
        ["https://example.com/dir1/dir2/resource/some-random-slug-3"]
    ],
    "max_attempts": 2
}'
```

As a result of this call, a batch execution request will be _started_.
The batch execution request is _not_ synchronous.

Example output:

```json
{
    "request_id":"a642f006-6631-4d96-9f84-171c5f96963e"
}
```

### Batch Execution Result REST API Call

Below is an example API request to get the result of a previously started Batch Execution.

```sh
curl --location '127.0.0.1:8080/api/v1/functions/execute/batch/result' \
--header 'Content-Type: application/json' \
--data '{
    "id": "d949d4e9-6419-4ad2-87c5-3a8deb33ee61"
}'
```

Example output:

```json
{
    "chunks": {
        "5302720e-d689-4463-adcb-8afb9c615135": {
            "peer": "12D3KooWNbW985igznpKwHAf1pxByyNeipQFtEFLynNrWhsEHkL9",
            "results": {
                "1f33048c02455bb49807ae58e2ccccca": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-13",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-13"
                    ]
                },
                "268a4145a50ade48aed2b1147d3518c6": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-1",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-1"
                    ]
                },
                "982931a535ee64f91fc822b5a0d3a555": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-9",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-9"
                    ]
                },
                "9954818207fe952736f370daf453f264": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-5",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-5"
                    ]
                },
                "fd37bb6de5f0b9daafdec0820a6fd349": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-17",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-17"
                    ]
                }
            }
        },
        "7397334c-54b6-4a15-8fab-824ec006a50e": {
            "peer": "12D3KooWH7RrtuB1YgXFB4piS776eyRrPfK34y2wL8tkTEwNoQfp",
            "results": {
                "2da1965d6a1239fa71e98fdab897ff8d": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-3",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-3"
                    ]
                },
                "ad69732dcf1756a2391fca4e8fd5c601": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-11",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-11"
                    ]
                },
                "b7396904551260cbc63ad6b6bf098bcf": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-19",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-19"
                    ]
                },
                "dbcb6ceb8e7a7c1e78743b8fb7629234": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-7",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-7"
                    ]
                },
                "fbbfa810e9c16122262f600f548594aa": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-15",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-15"
                    ]
                }
            }
        },
        "b9ce3ba3-004d-40b4-80f9-0cbf505d60f5": {
            "peer": "12D3KooWMTteLBk8fUtFVyh6FMFwKroGwek15fDc5bApdDxeVWTd",
            "results": {
                "2efac038c9d489a6f8057455b8cd9773": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-16",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-16"
                    ]
                },
                "4c555cef30403a7a11049c2883114da4": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-0",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-0"
                    ]
                },
                "52347f161caec8ccea34f1308d4ab3ab": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-4",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-4"
                    ]
                },
                "becb5828881f32bce44384c5c39b601b": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-8",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-8"
                    ]
                },
                "d5255aff17d6b4358e917fcf8ecc11b2": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-12",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-12"
                    ]
                }
            }
        },
        "e6ff1281-f10d-496b-895a-8056d4565a35": {
            "peer": "12D3KooWJHs23DwVANCvKHRrZwxWbLXn4gxBhZgsjh8B8VfiiwbZ",
            "results": {
                "42ec3ed349e3e3d029ddde64c7899c05": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-6",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-6"
                    ]
                },
                "8c7354c2a28bd99e0eef701234c7406e": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-2",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-2"
                    ]
                },
                "bbc6ceac22629ebcc3f2f5b0295360c0": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-18",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-18"
                    ]
                },
                "cc10dad585fb81bbb8822d030434d469": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-14",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-14"
                    ]
                },
                "e2a6032841c31e9d1dc5e73350d721ae": {
                    "result": {
                        "stdout": "https://example.com/dir1/dir2/resource/some-random-slug-10",
                        "exit_code": 0
                    },
                    "function_invocation": "bafybeie3nlygbnuxhvqv3gvwa2hmd4tcfzk5jtvscwl6qs3ljn5tknlt4q/echo.wasm",
                    "arguments": [
                        "https://example.com/dir1/dir2/resource/some-random-slug-10"
                    ]
                }
            }
        }
    },
    "code": "200",
    "request_id": "cb24c8cd-9c67-4bfc-8c87-2ce2ef4d8f10"
}
```

## Notes and Ideas for Improvements

### MongoDB ID types

We do not use the native MongoDB ID type for IDs for our records - batches, chunks or work items.
Instead, we use UUID v4 converted to strings.

This will be less performant and we can look into a different solution.

### MongoDB Indexes

Currently we do not have any indexes on MongoDB collections.

### Records in MongoDB queries are not split into parts

Currently we do not do any chunking when working with MongoDB - we do everything in a single query.
In case of inserting or updating very large number of work items, this may prove problematic and we split the input records into parts.

### Batches Are Resumed Sequentially

Because we're not sure on the number of available workers and other variables, when head node resumes batch execution, it will do this one by one.

### Work Items are Executed Sequentially by Workers

We could allow a degree of parallelism for Work Item processing.
However, in the scenario of web scraping e.g. 1000 URLs on a single domain, processing everything in parallel would likely lead to blocking of the offending worker.
For now, the safest option was to do this sequentially.
However, even now there is a risk of having too large rate of requests and still leading to bans.

### In-Progress Work Items

At the moment, a Work item could be assigned to a worker node and will be marked as `IN PROGRESS`.
A node could fail or something similar could happen and the Work item could remain stuck in this state.

We could (should) consider a mechanism to reset the state of the work item, to make it eligible for assignment to a different node.
This could be something like a time period during which work item has not been completed.
Alternatively (ties into the next point) we could ping the worker for the result and, if none is provided, reset the state of the work item then.

### Work Item Result Polling

Right now worker sends the results to head node after it has completed its chunk.
However, if a head node goes down temporarily, it might not receive it.

We might end up in a state where

1. Worker node is done with the execution, and
2. Head node did not receive the result, thus failed to persist it and update the DB.

We could introduce polling from the Head node to the worker node, effectively asking "do you have results of chunk `X` for me?".
If yes, it could update the information in the DB.
If not, we would need to differentiate between still in progress work items, and those that were "lost".
Worker nodes do not persist the information about work items they were assigned, so after restart worker would not be aware of the fact that something was expected from it.

### Number of Nodes Forced also on Batch Resume

For Batch Execution Request, we respect the number of nodes the user specified.
For example, the user might send Batch with 100 work items and request 10 nodes to process them.
These will be split into 10 chunks of 10 work items each.
Lets say one worker fails - we will have 90 `DONE` work items and 10 `FAILED`.

On Batch resume, we will pick up 10 non-complete work items, and resume batch execution.
However, we will again honor the number of nodes specified.
This means we will again request 10 nodes to process these 10 items.

We could be happy with a smaller number of workers and not force the original one.

### Multiple Identical Work Items in a Batch

Because of the way we calculate [work item IDs](#work-item-id), multiple work items in a single batch that are identical will produce identical IDs.
This might lead to collisions.
However, it did not seem like a good idea to manually validate uniqueness as batches could be quite large as it could be CPU intensive.
