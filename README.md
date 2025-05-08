# HealthChecks for your containers

[Docker Image](https://hub.docker.com/r/ewancoder/healthcheck)

Provides a binary that is able to perform the following healthchecks:

- Query a URI (send GET request)
- Check that file exists (and is up to date)

## The reason

Sometimes you want to include a healthcheck in some container, but it turns out the container is based on distroless image and doesn't have `ls` or `curl` or even `sh`. So you cannot execute shell commands in the container.

The solution is to include a custom binary into the image which we can specify as a healthcheck, the Docker will call this binary in order to determine whether the service is healthy or not.

## Usage

### Adding healthcheck binary

If you have a Dockerfile, add these lines to it:

```Dockerfile
# Copy Healthcheck executable.
COPY --from=ewancoder/healthcheck:latest /healthcheck /healthcheck
HEALTHCHECK CMD ["/healthcheck"]
```

This copies the latest version of the healthcheck executable and includes it into your image at `/healthcheck` location.

If you (like me) do not like to use third-party images for this, you can inspect the code to make sure there is nothing you dislike (it's very simple, and in a single file - check out `main.go`) and build it yourself:

```bash
git clone https://github.com/ewancoder/healthcheck
cd healthcheck
docker build -t myusername/myimagename .
docker push myusername/myimagename
```

This will effectively build and push the image to your own dockerhub repository, providing you the way to use your own image for healthchecking.

### Setting environment variables

You are required to specify one of the following environment variables for the healthcheck to work:

```env
HEALTHCHECK_URI=http://localhost:8080/health
HEALTHCHECK_FILE=/tmp/health
```

If you specify both of them - both healthchecks will happen at the same time and if one of them fails - healthcheck is considered unsuccessful.

#### URI healthchecks

If you specify `HEALTHCHECK_URI`, healthcheck tool will try to send a `GET` request to that URI and monitor for `200` status code.

You can control this behavior by specifying these (optional) environment variables:

```env
HEALTHCHECK_URI_TIMEOUT=10
HEALTHCHECK_URI_STATUS_CODE=200
```

Default values are as presented in the snipped above, timeout units are seconds.

#### File healthchecks

If you specify `HEALTHCHECK_FILE`, healthcheck tool will try to get the file at this location and verify that it has been changed in the last 60 seconds.

You can control this by specifying:

```env
HEALTHCHECK_FILE_MAX_AGE=60
```

Default value is as presented in the snipped above, units are seconds.

### Using the file healthcheck

Sometimes when you have an application that does not have any endpoints (a process, a tool, something else) - you need another type of health checking - like this file healthcheck.

However, in order for the healthcheck to consider your application healthy - you still need to create (and update) this file at least once per the `HEALTHCHECK_FILE_MAX_AGE` interval, so you need to add additional code to your application that will create this file in the scenario when you consider that it is healthy.

This is an example snippet in golang that does just that:

```go
func updateHealthCheckFile() {
	filePath := os.Getenv("HEALTHCHECK_FILE")
	if filePath == "" {
		fmt.Print("HEALTHCHECK_FILE environment variable is not set.")
		return
	}

	for {
		currentTime := time.Now().String()
		err := os.WriteFile(filePath, []byte(currentTime), 0644)
		if err != nil {
			fmt.Printf("Failed to write to health check file: %v\n", err)
		}
		time.Sleep(30 * time.Second)
	}
}

func main() {
	// Some initialization code
	// ...
	go updateHealthCheckFile();
}
```

Your application resides in the same container as the healthcheck application, so both will have access to the `HEALTHCHECK_FILE` environment variable. Your job is to read it, and to save/update the file at this location by a timer.
