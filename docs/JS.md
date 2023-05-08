# Server-Side Javascript

Taproot embeds the `goja` Javascript runtime, which allows you to execute arbitrary Javascript, either directly as 
server-side Javascripr (SSJS) or using the JSML markup language (see the JSML docs for more on this). Each SSJS script 
is run in a sandboxed JS runtime.

## JSCallReturnValue
Most calls to native functions return a complex object, `JSCallReturnValue`. Its fields are:
- `ok`: A boolean indicating whether the call succeeded.
- `errors`: An array of strings, representing errors (if any)
- `resultCode`: Result code, if applicable
- `resultDesc`: A description of a successful result, if applicable
- `results`: A JS object representing any successfully-returned data.

In general, you will want to test `result.ok` and then fetch data from `result.results`.

## Data Objects
The following objects are injected into JS and JSML runtimes:
- `req`: Request data
  - `req.host`: Hostname for the request
  - `req.method`: Request method
  - `req.url`: Request URL
  - `req.query`: Map of query objects
  - `req.body`: Body, if any
  - `req.form`: Form object, if applicable
- `context`: Maintains various context items from Taproot.
  - `context.realm`: The overall realm of the request
  - `context.domain`: The specific domain of the request
  - `context.user`: The user for the request
  - `context.rights`: An array of rights, if an Acacia policy has been applied and matched to the route.
  - `context.correlationId`: The tracing correlation ID for this request, also available at `correlationId`
  - `context.cspNonce`: The content security policy nonce, also available at `cspNonce`
  - `context.checkUserRight(userId, tenantId, rightId ,objectId)`: Checks to see if the stated user has a given right.
- `db`: Database-specific functions
  - `query(string dbName, string query, params...)`: Queries `dbName` with `query`, using `params` as SQL parameters.
    - Returns a `JSCallReturnValue` in which `data.rows` contains an array of JS objects, each one representing a database row.
  - `print()`: Prints a string to standard output, for debugging.
  - `dsns()`: Returns an array of strings, listing the various database IDs available.
- `data`: If any custom route-specific data is passed into this script, this is where it will appear.
- `util`: Utility functions
  - `print()`: Prints a string to the `deck` info log
  - `save(key, val)`: Saves a value to page storage. This and `export()` are useful to pass data to the JSML client side in a type-preserving way, particularly when using IDs (which overflow when not passed as a string).
  - `export()`: Exports the values interned using `save(k,v)` into JSON, for consumption on the client-side.