#!/usr/bin/env bash
set -xeuo pipefail

USER_NAME=$1

SDK_URL="https://dl.google.com/android/repository/sdk-tools-linux-4333796.zip"
ANDROID_HOME="/home/${USER_NAME}/android-sdk"

# Download Android SDK

echo ":: Install Android SDK"
echo ""
mkdir -p "$ANDROID_HOME" .android \
    && cd "$ANDROID_HOME" \
    && curl -o sdk.zip https://dl.google.com/android/repository/sdk-tools-linux-4333796.zip \
    && unzip sdk.zip \
    && rm sdk.zip \
    && yes | $ANDROID_HOME/tools/bin/sdkmanager --licenses

# Install Android Build Tool and Libraries
$ANDROID_HOME/tools/bin/sdkmanager --update

echo ":: Install api 27"
echo ""
# Install api 27
$ANDROID_HOME/tools/bin/sdkmanager "build-tools;27.0.3" \
    "platforms;android-27" \
    "platform-tools"
echo "Api 27: Install successfully"

echo ":: Install api 28"
echo ""
# Install api 28
$ANDROID_HOME/tools/bin/sdkmanager "build-tools;28.0.3" \
    "platforms;android-28" \
    "platform-tools"


echo ":: Set ANDROID_HOME to ENV"
echo ""

mv .bashrc .bashrc_original
echo -e "export ANDROID_HOME=$ANDROID_HOME\n" >> .bashrc
cat .bashrc_original >> .bashrc
rm .bashrc_original

echo ":: Successfully"
echo ""