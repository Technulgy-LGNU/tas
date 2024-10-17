# T.A.S. (Technulgy Admin Software)

This Software is for managing everything about the Technulgy organisation,
including but not limited to:
 - [ ] Members
 - [ ] Teams
 - [ ] Sponsors
 - [ ] Website
 - [ ] Newsletter
 - [ ] Orders
 - [ ] Events (Internal & External)

This software will include integrations for:
 - [ ] Email
 - [ ] Nextcloud
 - [ ] Discord
 - [ ] RCJ Forum Updates
 - [ ] RCJV Events and Game plans

Any ideas, email me (braunelias@tghd.email)

# Docker Compose
Only for Sysadmin (braunelias@tghd.email)

```yaml
# Docker Compose Production File

services:
  app:
    image: technulgy/tas:latest
    restart: always
    depends_on:
      - db
    environment:
      - DBUser=tas      
      - DBPassword=tasPassword    
      - Database=tas
      - TimeZone=Europe/Berlin 
      - InitialAdminUser=contact@technulgy.com 
      - InitialAdminPassword=1234     
      - EmailHost=mail.technulgy.com 
      - EmailApiKey=mhm
      - SenderEmail=noreply@tghd.email
      - SenderEmailPassword=2w@#i231&67D%9h$2T#6&z#cO@6@!!&0
      - NCHost=nc.technulgy.com
      - NCApiKey=mhm
    ports:
      - "80:80"
    volumes:
      - tas-data:/var/lib/tas/data

  db:
    image: postgres:16-alpine
    restart: always
    environment:
      - POSTGRES_USER=tas
      - POSTGRES_PASSWORD=tasPassword
    ports:
      - "5432:5432"
    volumes:
      - psql-data:/data

volumes:
  psql-data:
  tas-data:
```