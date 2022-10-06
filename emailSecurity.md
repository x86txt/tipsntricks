<!-- PROJECT SHIELDS -->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]


Protecting Domains that don’t send mail [^1]
------

If you have domains that don’t send mail, then it’s a good idea to protect does as well. This may sound strange, but these domains can still be used for spoofing and phishing attacks. You can also do this for subdomains that don’t send emails.  
By creating a simple DNS TXT record we can tell the receiving mail systems that mail from this domain is invalid and should be rejected.  

We can use a TXT record for this with the following format:  

> Name: *._domainkey.non-mail-domain.com  
> Value: v=DKIM1; p=

[^1]: [lazyadmin.nl](https://lazyadmin.nl/office-365/configure-dkim-office-365/)  




<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://shields.secunit.io/github/contributors/x86txt/prtg.svg?style=for-the-badge
[contributors-url]: https://github.com/x86txt/prtg/graphs/contributors
[forks-shield]: https://shields.secunit.io/github/forks/x86txt/prtg.svg?style=for-the-badge
[forks-url]: https://github.com/x86txt/prtg/network/members
[stars-shield]: https://shields.secunit.io/github/stars/x86txt/prtg.svg?style=for-the-badge
[stars-url]: https://github.com/github_username/repo_name/stargazers
[issues-shield]: https://shields.secunit.io/github/issues/x86txt/prtg.svg?style=for-the-badge
[issues-url]: https://github.com/x86txt/prtg/issues
[license-shield]: https://shields.secunit.io/github/license/x86txt/prtg.svg?style=for-the-badge
[license-url]: https://github.com/x86txt/prtg/blob/main/LICENSE
[linkedin-shield]: https://shields.secunit.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://www.linkedin.com/in/mevanssecurity/
[product-screenshot]: images/screenshot.png
