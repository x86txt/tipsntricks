# how to use nginx to serve your robots.txt file directly


```
    location = /robots.txt {
       add_header Content-Type text/plain;
       return 200 "User-agent: *\nDisallow: /\n";
    }
```
See [https://moz.com/learn/seo/robotstxt](https://moz.com/learn/seo/robotstxt) for more options.
