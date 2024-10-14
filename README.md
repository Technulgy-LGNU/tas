# T.A.S. (Technulgy Admin Software)

This Software is for managing everything about the Technulgy organisation,
including but not limited to:
 - Members
 - Teams
 - Sponsors
 - Website
 - Newsletter
 - Orders
 - Events (Internal & External)

This software will include integrations for:
 - Email
 - Discord
 - RCJ Forum Updates
 - RCJV Events and Game plans

# Docker Compose
Only for Sysadmin (braunelias@tghd.email)

```yaml
services:
  app:
    image: technulgy/tas:latest
    restart: always
    ports:
      - "80:80"
    depends_on:
      - db
    environment:
      - DBUser=showmaster      # Should correspond to the environment variables ->
      - DBPassword=password    # set in the db section of this docker compose file
      - Database=showmaster
      - TimeZone=Europe/Berlin # Change to your local timezone
      - AdminUserName=admin    # Email is admin@example.com
      - AdminPassword=1234     # Please change to a secure password

  db:
    image: postgres:16-alpine
    restart: always
    environment:
      - POSTGRES_USER=showmaster
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - psql-data:/var/lib/postgresql/data

volumes:
  psql-data:
```