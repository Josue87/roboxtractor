<h1 align="center">
  <b>roboXtractor</b>
  <br>
</h1>
<p align="center">
  <a href="https://golang.org/dl/#stable">
    <img src="https://img.shields.io/badge/go-1.16-blue.svg?style=flat-square&logo=go">
  </a>
   <a href="https://www.gnu.org/licenses/gpl-3.0.en.html">
    <img src="https://img.shields.io/badge/license-GNU-green.svg?style=square&logo=gnu">
  </a>
  <a href="https://github.com/Josue87/roboxtractor">
    <img src="https://img.shields.io/badge/version-0.2b-yellow.svg?style=square&logo=github">
  </a>
   <a href="https://twitter.com/JosueEncinar">
    <img src="https://img.shields.io/badge/author-@JosueEncinar-orange.svg?style=square&logo=twitter">
  </a>
</p>


<p align="center">
This tool has been developed to extract endpoints marked as disallow in robots.txt file. It crawls the file directly on the web and has a waybackmachine query mode (1 query for each of the previous 5 years).
</p>
<br/>

# 🛠️ Installation 

If you want to make modifications locally and compile it, follow the instructions below:

```
> git clone https://github.com/Josue87/roboxtractor.git
> cd roboxtractor
> go build
```

If you are only interested in using the program:

```
> go get -u github.com/Josue87/roboxtractor
```

**Note** If you are using version 1.16 or higher and you have any errors, run the following command:

```
> go env -w GO111MODULE="auto"
```

# 🗒 Options

The flags that can be used to launch the tool:

| Flag | Type | Description | Example |
|:----:|:----:|:------------|:--------|
| **u** | string | URL to extract endpoints marked as disallow in robots.txt file. | `-u https://example.com` |
| **m** | uint |  Extract URLs (0) // Extract endpoints to generate a wordlist (>1 default) | `-m 1` |
| **wb** | bool |  Check Wayback Machine. Check 5 years (Slow mode) | `-wb` |
| **v** | bool |  Verbose mode.  Displays additional information at each step | `-v` |
| **s** | bool |  Silen mode doesn't show banner | `-s` |

You can ignore the -u flag and pass a file directly as follows:

```
cat urls.txt | roboxtractor -m 1 -v
```

# 👾 Usage

The following are some examples of use:

```
roboxtractor --help
cat urls.txt | roboxtractor -m 0 -v
roboxtractor -u https://www.example.com -m 1 -wb
cat urls.txt | roboxtractor -m 1 -s > ./customwordlist.txt
cat urls.txt | roboxtractor -s -v | uniq > ./uniquewordlist.txt
echo http://example.com | roboxtractor -v
echo http://example.com | roboxtractor -v -wb
```
# 🚀 Examples

Let's take a look at some examples. We have the following file:

![image](https://user-images.githubusercontent.com/16885065/124949652-0bfb8c00-e012-11eb-83b7-2c4805570626.png)

Extracting endpoints:

![image](https://user-images.githubusercontent.com/16885065/125102251-5eed4600-e0db-11eb-9626-06ed1e5f7ec7.png)

Extracting URLs:

![image](https://user-images.githubusercontent.com/16885065/125102445-92c86b80-e0db-11eb-8406-ce0a87e3d78c.png)

Checking Wayback Machine:

![image](https://user-images.githubusercontent.com/16885065/125102688-db802480-e0db-11eb-8c2f-0d3b96733c00.png)

Github had many entries in the file, which were not useful, a cleaning process is done to avoid duplicates or entries with *. Check the following image:

![image](https://user-images.githubusercontent.com/16885065/125103171-62350180-e0dc-11eb-83fc-3b3b400690e8.png)


# 🤗 Thanks to 

The idea came from a tweet from [@remonsec](https://twitter.com/remonsec) that did something similar in a bash script. Check the [tweet](https://twitter.com/remonsec/status/1410481151433576449).
