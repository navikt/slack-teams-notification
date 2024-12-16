# Slack notification for NAIS teams
Send monthly notifications to Slack to keep NAIS teams up to date.

Fetch all teams and members from [NAIS API](https://github.com/nais/api):

```graphql
query teamsAndMembers {
    teams {
        nodes {
            slug
            members {
                nodes {
                    user {
                        name
                        email
                    }
                    role
                }
            }
        }
    }
}
```

For each member of the team, lookup the user from the Slack API to get the Slack handle, and send a formatted message to
the user. Only members with the `OWNER`-role for a team will receive notifications. 

## Configuration

Refer to the [internal/config](internal/config/config.go) package for the available configuration parameters for the job.