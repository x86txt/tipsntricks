
fail2ban filter for nginx deny module
------

> [!IMPORTANT]
> this guide assumes you have an enabled and working [deny rule](http://nginx.org/en/docs/http/ngx_http_access_module.html) in your nginx config


&nbsp;
&nbsp;


you will need to add a new fail2ban filter. below is where it should be located on debian/ubuntu, but other distributions may locate it somewhere else.

```shell
$ cat <<EOF | sudo tee /etc/fail2ban/filter.d/nginx-http-ipdeny1.conf 
# fail2ban filter configuration for nginx
# will catch any ips who fail nginx deny filter
# author: matthew evans

[Definition]

failregex = ^ \[error\] \d+\#\d+\: \*\d+ access forbidden by rule\, client\: <HOST>\,.+$
ignoreregex =
datepattern = {^LN-BEG}
EOF
```

```shell
$ sudo systemctl restart fail2ban
```

You can generate some test traffic from an IP that is not approved via a simple bash script like this:

```shell
#!/bin/bash

counter=1
while [ $counter -le 20 ]
do
    # make curl totally silent, but still return errors (-S)
    curl -S -s -o -4 ip.or.fqdn.of.your.fail2ban.machine
    ((counter++))
done
echo "Complete!"
```

After you generate some traffic and should have an IP that is banned, you can check with this command:

```shell
$ sudo fail2ban-client status nginx-http-ipdeny
```

finally, you can unban an IP with the following:

```shell
sudo fail2ban-client set nginx-http-ipdeny unbanip the.ip.to.unban.from.the.command.above
```
