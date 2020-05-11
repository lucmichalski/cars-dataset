Installs dependencies required to run 
[playwright](https://github.com/microsoft/playwright) 0.11.1.

Although Firefox runs fine, Webkit had a lot of issues
running which these dependencies fix.

At the time of testing, the versions of browsers were:

* chromium 740847
* firefox 1029
* webkit 1151

# Example

```
cd /example
npm run example
```

Check for screenshots at `/example/screenshots`.

If you want to view the screenshots on your container host machine,
you could run the image like:

```
docker run -it --rm \
  -v $(pwd)/screenshots:/example/screenshots \
  playwright-docker bash
```

Then open the screenshot files from the `screenshots` directory inside
your working directory.
