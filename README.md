# smsgw
Go based SMSGW

# How to Build
> go build

File called smsgw will be generated

# How to run
> ./smsgw

Access the smsgw portal at: http://yourhostnameorip:9083/smsgw

![image](https://user-images.githubusercontent.com/32011741/109934077-27116480-7cdd-11eb-8773-61a2c7881c45.png)

# Features
1. Multiple SMSC accounts backed with Goroutines.
2. Local & LDAP auth support.
3. Multiple sender IDs support per user group.
4. Role based portal access.
5. Customized Bulk SMS using simple template.

# ToDO
1. Reporting
2. Split UI handler & smsgw core using lightweight messaging solution .e.g NATS or REDIS
3. etc.
