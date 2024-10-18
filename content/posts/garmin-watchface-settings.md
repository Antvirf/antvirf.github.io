+++
author = "Antti Viitala"
title = "Adding settings to a custom Garmin watch face"
date = "2022-10-21"
description = "Guide on adding settings and properties to a watch face to provide your users with more personalization options."
tags = [
    "development"
]
thumbnail = "/content/hero.jpg"
images = ["/content/hero.jpg"]
series = []
+++

## Objective

This article is an incremental set of changes to the "Matter Metrics" watch face that was created in a previous tutorial [here](https://aviitala.com/posts/garmin-watchface-tutorial/). The following features will be added:

* Ability to customize date format: User can choose from 4 available types of date info, separated by dash. These are, using 22nd of October as an example:
  * ```mm``` (10)
  * ```mmm``` (Oct)
  * ```dd``` (22)
  * ```ddd``` (Sat)
* Ability to customize start and end time for the time gauge at the top of the watch face. As Garmin doesn't provide an API for sunrise and sunset times, these will have to be user configurable instead.

## Basic structure and concepts of a setting

### ```property```

A ```property``` of a watch face can be thought of as a configuration option that can be defined outside the Monkey C code itself. In order to bring in configurable values, each value to be brought in must have its own ```property``` defined in the ```resources/properties.xml``` file. A ```property``` does not *need* to be attached to a user-customizable setting; the developer could add certain properties outside the source code for easier maintenance.

A ```property``` is defined as shown below, and comes with an ```id```, a ```type``` (for available types see [the docs](https://developer.garmin.com/connect-iq/core-topics/properties-and-app-settings/)) and a default value as shown below.

```xml
<properties>
    <property id="TimeGaugeStartValue" type="number">1</property>
    ...
</properties>
```

Accessing a ```property``` from within the watch face code is straightforward - use the following snippet, and replace the shown string value with the ```property.id``` used in the ```properties.xml```:

```c
Application.getApp().Properties.getValue("TimeGaugeStartValue");
// would return 1 by default
```

### ```setting``` and ```settingConfig```

In order for the end user to be able to change the ```property```, the project must attach a ```setting``` to declare that it can be modified, as well as a ```settingConfig``` to describe the options available to the user - and what those options correspond to in code. This information is used by ConnectIQ to create the settings menu for your watch face.

Below is the example menu options available for the time gauge start value. the ```value``` corresponds to what gets passed to the program, while the string value inside the tags (e.g. "0600") is the value shown to the user for that particular option.

```xml
<settings>
    <setting propertyKey="@Properties.TimeGaugeStartValue" title="@Strings.TimeGaugeStartString">
        <settingConfig type="list">
            <listEntry value="0">"0600"</listEntry>
            <listEntry value="1">"0630"</listEntry>
            <listEntry value="2">"0700"</listEntry>
            <listEntry value="3">"0730"</listEntry>
            <listEntry value="4">"0800"</listEntry>
            <listEntry value="5">"0830"</listEntry>
            <listEntry value="6">"0900"</listEntry>
            <listEntry value="7">"0930"</listEntry>
            <listEntry value="8">"1000"</listEntry>
            <listEntry value="9">"1030"</listEntry>
            <listEntry value="10">"1100"</listEntry>
            <listEntry value="11">"1130"</listEntry>
            <listEntry value="12">"1200"</listEntry>
        </settingConfig>
    </setting>
    ...
</settings>
```

Other ```settingConfig``` types are also available - for a simple on/off toggle, use ```boolean``` similar to the military time setting:

```xml
<setting propertyKey="@Properties.UseMilitaryFormat" title="@Strings.MilitaryFormatTitle">
    <settingConfig type="boolean" />
</setting>
```

### ```string```

You may already have noticed the previous snippets contained some annotations like ```"@Strings.MilitaryFormatTitle"```. The ```strings.xml``` allow you to separate the actual text that the user sees from the rest of the implementation.

This does not seem to be mandatory, in my experience so far any value that contains "@Strings.*" could just be replaced directly with the actual text you want to show to the user.

## Date format setting

The commit with the full changes can be found [here](https://github.com/Antvirf/garmin-watch-face-guide/commit/236d7be3b277eef08f73c600d46aeaee76f0511d). It is best to create a separate function to convert a given settings value into your desired state in the code, as shown in the example below with ```getDateInfoAsString()```.

```c
private function getDateInfoAsString(dc as Dc, option as Number) as String {
    var info = Gregorian.info(Time.now(), Time.FORMAT_SHORT);
    var longInfo = Gregorian.info(Time.now(), Time.FORMAT_LONG);
    switch (option){
        case 0: // Day of month, number
            return Lang.format("$1$", [info.day]);
        case 1: // Month of year, number
            return Lang.format("$1$", [info.month]);
        case 2: // Month of year, text
            return Lang.format("$1$", [longInfo.month]);
        case 3: // Day of week, text
            return Lang.format("$1$", [longInfo.day_of_week]);
    }
```

## Time gauge setting

The commit with the full changes can be found [here](https://github.com/Antvirf/garmin-watch-face-guide/commit/cb79f42f66411ba20838da754c4c2e87bfd0a483). Given the larger number of cases and options for the time gauge start and end settings, I wrote a short Python script to generate the XML lines as well as the switch-case statements. The function itself is relatively simple as can be seen below:

```c
// Get corresponding start time value, given app setting
private function getStartTimeValue(dc as Dc, option as Number) as Array<Number>{
    switch (option){
        case 0:
            return [6,0];
        case 1:
            return [6,30];
        case 2:
            return [7,0];
        // ... other case options
    }
```

The flow is roughly as follows:

1. Read the properties value
1. Convert the single numeric property value into an array of [hours, minutes]
1. Split the array into hours and minutes separately for computations, pass on to time gauge function
1. Take the array and prettify it into a string to be shown to the user - **taking into account military time setting**

## References

To learn more, check out [this page on properties and app settings](https://developer.garmin.com/connect-iq/core-topics/properties-and-app-settings/) from the Garmin SDK docs.
