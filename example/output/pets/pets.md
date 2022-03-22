# Pets

Pets are excellent friends. Why not get one via the API.


## Create Pet

 - server: https://animals.example.com"
 - summary: Create Pet
 - method: post
 - path: /pets
 - queryParams: W10=
 - headerParams: W3sibmFtZSI6IkF1dGhvcml6YXRpb24iLCJyZXF1aXJlZCI6dHJ1ZSwiZGVzY3JpcHRpb24iOiJUaGUgYmFzaWMgYXV0aG9yaXphdGlvbiBoZWFkZXIgdG8gYXV0aG9yaXplIGFnYWluc3QgdGhlIEFQSSIsInR5cGUiOiJzdHJpbmciLCJleGFtcGxlIjoiQXV0aG9yaXphdGlvbjogQmFzaWMgUVZCSlgwdEZXVG89In0seyJuYW1lIjoiQ29udGVudFR5cGUiLCJyZXF1aXJlZCI6dHJ1ZSwiZGVzY3JpcHRpb24iOiJUaGUgcmVxdWVzdCBjb250ZW50IHR5cGUiLCJ0eXBlIjoic3RyaW5nIiwiZXhhbXBsZSI6IkNvbnRlbnQtVHlwZTogYXBwbGljYXRpb24vanNvbiJ9XQ==
 - requestParams: W3sibmFtZSI6Im5hbWUiLCJkZXNjcmlwdGlvbiI6IlRoZSBuYW1lIG9mIHlvdXIgcGV0IiwidHlwZSI6InN0cmluZyIsImV4YW1wbGUiOiJHYXJmaWVsZCIsImVudW0iOm51bGwsInJlcXVpcmVkIjp0cnVlfSx7Im5hbWUiOiJ0eXBlIiwiZGVzY3JpcHRpb24iOiJUaGUgdHlwZSBvZiBwZXQgeW91IHdhbnQiLCJ0eXBlIjoic3RyaW5nIiwiZXhhbXBsZSI6ImNhdCIsImVudW0iOm51bGwsInJlcXVpcmVkIjpmYWxzZX1d
 - title: Create Pet
 - description: Create a new pet
 - responseExampleKeys: success

## Curl command:

```bash
curl -X post 'https://animals.example.com/pets' \
	-H 'Authorization: Basic QVBJX0tFWTo=' \
	-H 'Content-Type: application/json'
	-d@- <<EOF
	{
        "name": "Garfield",
        "type": "cat"
    }
EOF
```

## Responses:

The pet has been created

```json
{
    "name": "Garfield",
    "type": "cat"
}
```


