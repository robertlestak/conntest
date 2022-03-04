# conntest

A single-binary cross-platform lightweight client/server connection testing tool. Currently supports L7 TCP (HTTP).

## Configuration Options

- `concurrency`: The number of concurrent connections to make
- `description`: The description of the test
- `run_count`: The number of times to run the test
- `client_delay_ns`: Optional client-side delay to introduce, in nanoseconds
- `server_delay_ns`: Optional server-side delay to introduce, in nanoseconds
- `upstream_endpoint`: Optional endpoint upstream to server, which server will `GET` on each client request
- `upstream_timeout_ns`: Optional timeout for upstream requests, in nanoseconds. Default is infinity.
- `data`: Optional data to send to the server. Default is `uuid`.

## Usage

### Remote Server

Start `conntest` in `server` mode on the remote instance.

```bash
conntest server -p 8080
```

### Local Client

Make a request to the server to create a test plan.

```bash
# set CONNTEST remote endpoint
export CONNTEST=http://conntest-server:8080
# create a new test group, storing the response in a variable
RESP=`curl $CONNTEST/test-groups/create -d '{
	"description": "example test group",
	"run_count": 1000,
	"concurrency": 5,
	"client_delay_ns": 100,
	"upstream_endpoint": "https://internal.example.net",
	"server_delay_ns": 0
}'`
# export the run_group_id from the response JSON
export run_group_id=`echo $RESP | jq -r '.run_group_id'`
```

The server will respond with a `run_group_id`.

Now, start `conntest` in `client` mode, providing the remote server endpoint, the run group ID to run, and the path where to generate the report file.

```bash
conntest client -r $CONNTEST -g $run_group_id -f report.json -d "$(<data.txt)"
```

## Results

Upon completion of a test, `conntest` will iterate over the client and server results to determine if there are any discrepancies in the data sent by the client and reported received by the server. If there are discrepancies, the test will fail. The average round trip time (as seen from the client perspective) is reported, as well as the raw round trip times for each request.

After completing a test run, the client will store the results in a local JSON file. 

### Analyzing Results

#### Get Average Response Time (in nanoseconds)

```bash
REPORT_FILE=report.json
jq '.average_response_time_ns' $REPORT_FILE
```

#### Find Any Errors

```bash
jq '.results[] | select(.error != null)' $REPORT_FILE
```

#### Get All Durations

```bash
jq '.results |=sort_by(.run_count)|.results[]|.client_duration_ns' $REPORT_FILE
```

#### Create Graph of Durations

```bash
export SCRIPTS_DIR=scripts
export REPORT_FILE=report.json
bash scripts/plot.sh
open roundtrip-plot.png
````