# smsgw
Go based SMSGW

# How to Build
> go build

File called smsgw will be generated

# How to run
> ./smsgw

Access the smsgw portal at: http://yourhostnameorip:9083/smsgw

![image](https://user-images.githubusercontent.com/32011741/109936773-4ad5aa00-7cdf-11eb-8050-ac0c9e1f29ae.png)

# Features
1. Multiple SMSC accounts backed up with Goroutines.
2. Local & LDAP auth support.
3. Multiple sender IDs support per user group.
4. Role based portal access.
5. Customized Bulk SMS using simple template.
6. Multiple smpp connections per account.

# ToDO
1. Reporting
2. Split UI handler & smsgw core using lightweight messaging solution .e.g NATS or REDIS
3. etc.
