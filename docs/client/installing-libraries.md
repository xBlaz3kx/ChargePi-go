# 💻 Installing libraries

## Read before continuing

The process of installing the dependencies was automated by creating a [script](../../scripts/install-dependencies.sh) for
installing the necessary dependencies. The script has the following arguments:

| Argument | Default value |                                                    Description                                                    |
|:--------:|:-------------:|:-----------------------------------------------------------------------------------------------------------------:|
|    0     |  "pn532_i2c"  | The libnfc configuration for PN532 RFID card reader. Possible values are: `pn532_uart`, `pn532_i2c`, `pn532_spi`. |
|    1     |   false (0)   |                         Whether or not to install Go on the system using a shell script.                          |

Example usage:

```bash
 cd ~/ChargePi-go/docs
 chmod +x install-dependencies.sh
 ./install-dependencies.sh pn532_i2c 1
```

## Building libnfc for PN532

1. Get and extract the libnfc:

    ```bash
     cd ~
     mkdir libnfc && cd libnfc/
     wget https://github.com/nfc-tools/libnfc/releases/download/libnfc-1.8.0/libnfc-1.8.0.tar.bz2
     tar -xvjf libnfc-1.8.0.tar.bz2
    ```

   **Next two steps may vary for your reader and communication protocol**

2. Create PN532 configuration:

    ```bash
     cd libnfc-1.8.0
     sudo mkdir /etc/nfc
     sudo mkdir /etc/nfc/devices.d
     sudo cp contrib/libnfc/pn532_uart_on_rpi.conf.sample /etc/nfc/devices.d/pn532_uart_on_rpi.conf 
     sudo nano /etc/nfc/devices.d/pn532_uart_on_rpi.conf
    ```

3. Add the line at the end of the file:

    ```text
    allow_intrusive_scan = true
    ```

4. Install dependencies and configure:

    ```bash
     sudo apt-get install autoconf
     sudo apt-get install libtool
     sudo apt-get install libpcsclite-dev libusb-dev
     autoreconf -vis
     ./configure --with-drivers=pn532_uart --sysconfdir=/etc --prefix=/usr
    ```

5. Build the library:

    ```bash
    sudo make clean
    sudo make install all
    ```

## Installing rpi-ws281x library

Follow the instructions from the [maintainer's repository](https://github.com/jgarff/rpi_ws281x).

**TLDR; Use the instructions here**

```bash
git clone https://github.com/jgarff/rpi_ws281x
cd rpi_ws281x
mkdir build
cd build
cmake -D BUILD_SHARED=OFF -D BUILD_TEST=ON ..
cmake --build .
sudo make install
```

To be able to use this C library in Go, it must be installed. You can do this by copying `*.h` to `/usr/local/include`
and `'.a` files to `/usr/local/lib`. If not working, check the Go wrapper library
instructions [here](https://github.com/rpi-ws281x/rpi-ws281x-go).
