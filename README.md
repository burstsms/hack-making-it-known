# hack-making-it-known
This is the repo for our hackathon AI slack bot project

## Miro Board
https://miro.com/app/board/uXjVMG9P-hE=/?share_link_id=518528715956

## Slack App

## Cloud Functions
There are two cloud functions. 
HTTP and PubSub. 
The HTTP function is used to handle the slack events and publish pertinent 
events to the pubsub topic.
The PubSub function is used to handle the pertinent slack events by passing the 
message text to OpenAI as a prompt, and then posting the completion to the Slack channel.
Done this way in-case OpenAI does not respond before the Slack event request times out.

## Development dev_http main.go
The main.go in the dev directory is used to run the slack event receiver locally. 
When deployed to cloud function the Handler method is the entry point.

## HTTP Environment Variables
The following environment variables are required to run the bot locally or in cloud function.
- SLACK_SIGNING_SECRET: 
  - the signing key we get for the MIK Slack app used to verify the that the event came from our Slack bot.
  - https://api.slack.com/authentication/verifying-requests-from-slack
  - if omitted the slack event receiver will not verify the request came from Slack
- CLOUD_PUB_SUB_TOPIC:
  - Used to specify the Cloud PubSub event topic
  - Defaults to `hack-slack-bridge`
- CLOUD_PROJECT_ID:
  - Used to specify the Cloud Project ID
  - Defaults to `tmp-hack-no-team-name`

Running locally you need:
- FUNCTION_TARGET: 
  - Used to specify the registered entrypoint name
  - Needs to be `MakeItKnown` 

## PubSub Subscriber Environment Variables
The following environment variables are required to run the bot locally or in cloud function.
- OPENAI_API_KEY: the OpenAI API key
- SLACK_BOT_TOKEN: the Slack bot token
- OPENAI_MODEL: the OpenAI model to use

## Deployment

### Cloud Function

#### Prerequisites


