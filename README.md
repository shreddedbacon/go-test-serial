# Backplane Controller Service (spackler)

This runs on the raspberry pi that has the backplane controller pi HAT on it

To run the service, start the docker container so it can talk to the raspberry pi serial port that is connected to the pi HAT
```
docker run -it --rm --device /dev/ttyS0 -p 8585:8585 -e SERIAL_DEVICE="/dev/ttyAMA0" spacklerind/spackler
docker run -it --rm --device /dev/ttyS0 -p 8585:8585 -e SERIAL_DEVICE="/dev/ttyS0" spacklerind/spackler
```

## Interacting with it
```
curl -X GET {target}/api/v1/power/{i2cAddress}/{i2cSlot}/{powerStatus}
```
