## high-level security precautions for various cloud environments and products

### Common Usage, i.e. "Security 101"
- always put in the effort to minimally scope account, policy, role, or token privilege (principle of least privilege)[^1]
- create a limited USER in your Dockerfile and switch to it with the USER command[^3].
  ```docker
  RUN groupadd -r appuser -g 433 && \
      useradd -u 431 -r -g appuser -s /sbin/nologin -c "Docker image user" appuser
      
  USER root
  RUN  somecommand.sh     # command that needs root
  RUN  anothercommand.sh  # another command that needs root
  
  USER appuser           # switch back to the limited user
  ```

### AWS

#### ALB
- only allow traffic to ALB from Cloudfront or Cloudflare
  - from Cloudflare: can use simple Lambda to scrape https://www.cloudflare.com/ips/
  - from Cloudfront: https://aws.amazon.com/blogs/networking-and-content-delivery/limit-access-to-your-origins-using-the-aws-managed-prefix-list-for-amazon-cloudfront/
    - use IPs in NACLs as backup

#### Fargate 
- create security group for cluster service that only allows traffic to the service from the ALB security group

#### ECR
- enable tag immutability to prevent images being overwritten


### Github   

- use any of the free code scanning SAST tools, see [sast-scan.yml](https://github.com/x86txt/tipsntricks/blob/5d8a801a86b7777b6406e073e228a841cd0e3af2/samples/sast-scan.yml) for example of a ready-to-go Github Action
- use CODEOWNERS to prevent senstive files or actions being overwritten without approval  
  - https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners  

- Github Actions[^3]:  
  - don't use structured data (i.e. JSON, XML, YAML) in secrets
  - use intermediate environment variable for untrusted input
    - accept the input into a variable, then use THAT variable in your sensitive function

[^1]: [https://en.wikipedia.org/wiki/Principle_of_least_privilege](https://en.wikipedia.org/wiki/Principle_of_least_privilege)  
[^2]: [Docker USER reference](https://docs.docker.com/engine/reference/builder/#user)
[^3]: [Github's Official Guidance](https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions)
