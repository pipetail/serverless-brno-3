import { SSTConfig } from "sst";
import { Api, WebSocketApi, Table, Function, Queue } from "sst/constructs";

export default {
  config(_input) {
    return {
      name: "notification",
      region: "eu-west-1",
    };
  },
  stacks(app) {
    app.setDefaultFunctionProps({
      runtime: "go1.x",
    });

    app.stack(function Stack({ stack }) {

      // queues
      const notifyConnection = new Queue(stack, "notifyConnection");
      const notifyUser = new Queue(stack, "notifyUser");
      
      // REST api
      const api = new Api(stack, "api", {
        routes: {
          "POST /token": "cmd/token/issuer/main.go",
        },
      });

      // persistence for websockets
      const userIdIndexName = 'UserIdIndex';
      const connections = new Table(stack, "connections", {
        fields: {
          ConnectionId: "string",
          UserId: "string",
        },
        primaryIndex: { partitionKey: "ConnectionId" },
        globalIndexes: {
          [userIdIndexName]: {
            partitionKey: "UserId",
            projection: "keys_only",
          },
        },
      });

      // websocket api
      const wsApi = new WebSocketApi(stack, "wsapi", {
        routes: {

          // execute when connection is opened
          $connect: {
            function: {
              timeout: 10,
              handler: "cmd/connect/main.go",
              permissions: [connections],
              environment: {
                CONFIG_CONNECTIONS_TABLE_ID: connections.tableName,
              },
            }
          },

          // execure when connection is closed
          $disconnect: {
            function: {
              timeout: 10,
              handler: "cmd/disconnect/main.go",
              permissions: [connections],
              environment: {
                CONFIG_CONNECTIONS_TABLE_ID: connections.tableName,
              },
            }
          },

          // handle messages that don't match the pattern
          $default: {
            function: {
              timeout: 10,
              handler: "cmd/default/main.go",
              permissions: [notifyConnection],
              environment: {
                CONFIG_SQS_NOTIFY_CONNECTION_URL: notifyConnection.queueUrl,
              },
            }
          },

          // handle specific messages
          ping: {
            function: {
              timeout: 10,
              handler: "cmd/ping/main.go",
              permissions: [notifyConnection],
              environment: {
                CONFIG_SQS_NOTIFY_CONNECTION_URL: notifyConnection.queueUrl,
              },
            }
          },
        },
      });

      // notify connection  consumer      
      notifyConnection.addConsumer(stack, {
        function: {
          timeout: 10,
          handler: "cmd/notify_connection/main.go",
          permissions: [wsApi],
          environment: {
            CONFIG_API_GATEWAY_ENDPOINT: wsApi.url.replace("wss://", "https://"),
          },
        }
      });

      // notify user queue / consumer
      notifyUser.addConsumer(stack, {
        function: {
          timeout: 10,
          handler: "cmd/notify_user/main.go",
          permissions: [notifyConnection, connections],
          environment: {
            CONFIG_CONNECTIONS_TABLE_ID: connections.tableName,
            CONFIG_SQS_NOTIFY_CONNECTION_URL: notifyConnection.queueUrl,
            CONFIG_USER_ID_INDEX_NAME: userIdIndexName,
          },
        }
      });

      // set console outputs
      stack.addOutputs({
        ApiEndpoint: api.url,
        WsApiEndpoint: wsApi.url,
      });

    });
  },
} satisfies SSTConfig;
