# Slack notification for NAIS teams
Send monthly notifications to Slack to keep NAIS teams up to date.

Fetch all teams and members from the [teams-backend](https://github.com/nais/teams-backend) GraphQL API:

```graphql
{
    teams {
        slug
        members {
            user {
                name
                email
            }
            role
        }
    }
}
```

for each member of the team, lookup the user from the Slack API to get the Slack handle, and send a formatted message to 
the user. Only members with the `OWNER`-role for a team will receive notifications. 