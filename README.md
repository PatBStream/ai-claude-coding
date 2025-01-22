# Example Using Claude AI for Software Design and Coding

Using Claude AI Chat, to design and code a complete application, from specific requirements.

**This prompt was used in Claude for this app:**
```console
Write a complete GOLANG app to run as a server to process millions of connections from clients, 
sending and receiving JSON messages.  Use best practices and robust error handling to ensure stability 
and reliability.  Also include bash scripts to build and run the app.

create a Docker compose script, using best practices for Docker and deployment, 
to build a container for this app and run it on Docker, listening on port 28999
```

The GOLANG code was complete and compiled.  The Docker build failed with errors and the prompt to Claude to fix:
```console
the Docker compose command fails with errors:

WARN[0000] /home/pat/projects/aiagent/claude/compose.yml: the attribute version is obsolete, it will be ignored, please remove it to avoid potential confusion
[+] Running 0/1
```
Next, updated the code to process JSON messages.  The prompt to update the code:
```console
update the server code to extract the JSON message fields and print them out.
```
Claude updated the code for this requirement.  Additionally, it provided this JSON message for testing the app.
```console
{
    "id": "msg123",
    "type": "data",
    "source": "client1",
    "payload": {
        "name": "test",
        "value": 42,
        "nested": {
            "field1": "value1"
        }
    }
}
```
To use this JSON message to test the app:
```console
Command to test using netcat (nc)
echo '{"id":"msg123","type":"data","source":"client1","payload":{"name":"test","value":42,"nested":{"field1":"value1"}}}' | nc -q 0 localhost 28999
```
