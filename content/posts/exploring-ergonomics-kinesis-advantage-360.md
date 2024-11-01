+++
author = "Antti Viitala"
title = "Exploring ergonomics and learning to type again with Kinesis Advantage 360"
date = "2024-11-01"
description = "My experiences with various keyboards and mice, and notes on learning to use a strangely-shaped keyboard in the form of the Kinesis Advantage 360 (split, columar and keywelled)."
tags = [
    "ergonomics",
]
images = ['content/kinesis-full.jpeg']
+++

![desk](/content/kinesis-full.jpeg)

*Links to products on this page may be affiliate links.*

## Earlier experiments

After using computers, keyboards and mice for only about two decades, I started to experience pain around my wrists and forearms, especially badly after long gaming sessions. Since I wasn't going to stop using computers, I had to start looking into this problem, and I started with keyboards. My work consists mostly of writing (code, docs, emails, messages) along with occasional mouse use (using online tools, dashboards etc.).

Before the [Kinesis Advantage 360](https://amzn.to/3NNe9ih), I've had, in rough order:

1. 'Standard' QWERTY keyboards of a few types, usually with mechanical switches. Normal layout, straight keyboard. Theses in the end gave me a fair amout of wrist pain on *both* hands, which by my guess is a result of ulnar deviation (my shoulders are much wider than the position of my wrists), causing wrist pain - example picture from Kinesis below:
  ![ulnardeviation](/content/kinesis-ulnar-deviation.webp)
2. Based on this hypothesis, my next keyboard was the Microsoft Sculpt shown below - note the slight split in the middle (which would hopefully bring my hands further apart, more in line with the shoulders) and the slight tilt/wave of the keypad (which would hopefully help keep wrists straight instead of having to bend to the angle of the board). The keyboard took maybe a week at most to get used to. In terms of ergonomics, it was a major improvement - this change eliminated all pain from the **left** wrist/arm. The right hand improved, but I could never get used to mouse and the pain didn't go away completely.
  ![mssculpt](/content/microsoft-sculpt.webp)
3. After about a year, the Sculpt keyboard broke down and I replaced it with the [Logitech Ergo K860](https://amzn.to/4hqq3fK). The K860 was much higher quality and would last for years being carried around without a problem. It had the same benefits due to its slightly split and 'wavy' layout, which already helps a lot without needing to learn how to type again from scratch.
  ![k860](/content/ergo-k860.png)
4. I saw a used DIY-built Dactyl Manuform build for sale locally, and decided to give it a try. The layout of the board was [colemak](https://colemak.com/) which I did not bother changing. I gave this a few weeks alongside work to try to learn it but it proved too difficult - trying to learn a new layout and a new keyboard shape at the same time was not a great decision in hindsight, and the 3D-printed case of this board made it too light, so it moved around a lot on the desk. Complete lack of wrist support also made this uncomfortable. Gave up after a few weeks.
  ![dactyl](/content/dactyl-manuform.jpeg)
5. At this point, I turned my focus to mice, and tried a few alternatives:
  - Previously, various gaming mice like Zowie FK, Logitech G502. All of them cause similar pain.
  - [Logitech MX Ergo trackball](https://amzn.to/4hqjA4u): Felt a little bit better at first, and then a whole lot worse after a few days. Made my hand hurt worse than before due to having to use the thumb so much to move the ball. Perhaps good if you use your mouse *very* little (or for very small motions only), but not workable for me. Returned after a couple of weeks switching back and forth with a regular mouse.
  - [Apple Magic Trackpad](https://amzn.to/3UtFmuh): Figured it would be worth a try but caused sharp pains after a couple of days of use.
  - [Logitech MX Master 3](https://amzn.to/3UyF5q2): Worst pain out of everything I have tried, returned it in a matter of days.
  - [Logitech MX Vertical](https://amzn.to/3YuuMUP): Trades off a lot of precision but does reduce pain. Not good for extended mouse use, but works well to quickly take care of something before getting back to the keyboard. This is what I've stuck with for now.

Finally after several years of exploration, the stars aligned in the form of someone selling their barely used Advantage 360. which I picked up before going on leave for a couple of weeks and promised myself to at least get decent with it.

## Few bullet points on first impressions

- The keyboard is *very sturdy*, and does not move around on the desk at all.
- It is quite difficult to carry around due to its weight and strage shape (though this tradeoff is worth it for how high quality it feels).
- MX brown switches take more force than I would prefer - after using mainly laptop and membrane keyboards at work for some time, the heavier feel took some getting used to.
- The SmartSet configuration program works well for everything I have had to do with it. The "Pro" version with Bluetooth comes with more modding capability, but given there have been some complaints about connection stability etc., I was happy with the more limited configurability. Basic macros and keybinds can be done on the keyboard itself without any external apps which is nice.

## Months 1-2: Struggle is real

### Practice, practice, practice

There are countless Reddit threads discussing the steep learning curve of Kinesis Advantage keyboards, and they are not wrong. The first few weeks feel incredibly awkward as your brain tries to adapt itself to the new shape and layout. The feeling is similar to switching your mouse hand, or chopping vegetables with your non-dominant hand. During the first couple of weeks of use I was on leave, and would do typing exercises a few times a day - usually 2-5x 1-minute runs of typing as a "set", and then 1-3 sets a day.

For practice, even though I kept my layout as standard QWERTY, I used the [colemak.academy](https://colemak.academy/) for practice (make sure to turn *off* keyboard mapping in the lower right corner). The practice runs are always 50 words, and give you an overall WPM rate at the end. My peak WPM on the K860 was about 115, so I set my goal to get to 100 WPM with the Kinesis.

On the first days, I started at 5-10 WPM. The progression from that point on, for the first two months, looked something like this:

- Week 1: 10-20 WPM
- Week 2: 20-30 WPM
- Week 3: 30-50 WPM
- Week 4: 50-70 WPM
- Weeks 5-8: 60-80 WPM

The first week I only used the keyboard during typing practice, as using it for normal tasks was exhausting and slow. After the first week, I started using it 'regularly', though at times I would still need to move it aside to get something quick done.

### Encouraging the use of keyboard shortcuts

Due to the shape of the keyboard, your hands sit slightly higher. Due to its split nature, likely your hands are spaced further apart from each other. As a result, the mouse moves further away on the desk from your body than you are probably used to, which makes reaching for it more of a chore. Being slow to type on a new keyboard, the feeling of extra effort reaching for the mouse is also amplified, as it further slows down what you are doing.

This adds what can be considered a "good" level of friction to your mouse usage. Since reaching for the mouse takes time and effort, I found myself looking for, learning, and using keyboard shortcuts for almost all applications much more frequently. For occasional minor mouse usage, [`warpd`](https://github.com/rvaiya/warpd) allows you to control a virtual mouse pointer with your keyboard which can come very handy. Bind the shortcut to a macro key, and you can use a virtual pointer with a single button press.

### Upgrades and minor mods

Once I started getting past 30 WPM or so, I began noticing that the keyboard was rather loud. To quiet it down, I got some [o-rings intended for mechanical keyboards](https://amzn.to/4ebL9eO) to make the keyboard less annoying and more office-friendly. O-rings tend to take a couple of weeks to reach their stable state, so even though initially the switches felt 'mushy' after this change, that disappeared over time while the benefits of reduced noise remain.

The placement of your hands on the keyboard takes some getting used to as it is very easy to inadvertedly keep your wrists too low, causing pain. The [wrist pads Kinesis provides](https://amzn.to/40oM3By) are good quality, and I use them about ~50% of the time when using the keyboard, as they are quick and easy to put on and take off when you need to.

### 'Proper' typing

I had learned to type with a regular keyboard naturally and never spent any time analysing how I typed, or otherwise trained to improve it. I used index and middle fingers for probably 80% of keystrokes, leaving ring finger and pinky only only for modifiers. The columnar layout used by the Kinesis Advantage 360 forces you to fix this - each finger has their own columns of keys to address, and using it in any other way feels wrong. For me, learning to type on the Kinesis was about learning to type 'properly' for the first time. Even when switching back to regular keyboards or using a laptop, this change of using fingers more evenly actually persists, which is likely beneficial for the long run even when not using the Kinesis itself.


## Months 3-6: Awkwardness and friction

At this point, I had reached peaks of ~90 WPM, though at the 3-month mark my consistent score was around 60-70. Writing symbols was still significantly slower compared to the K860.

### Barrier between you and the machine

I had by now reached typing speeds that more or less matched my general speed of thought, so I was surprised to notice that I frequently had the desire to 'just throw the weird keyboard aside for a second' when faced with a problem I generally knew how to solve already. The weirdly shaped keyboard was still something I was consciously aware of using, and I would occasionally move it aside to quickly sprint through a problem.
This feeling of the keyboard introducing friction took longer than 6 months dissipate. In hindsight this was probably due to the typing exercises focusing solely on letters - no numbers or symbols - so the amount of practice I had writing symbols which are very common in code was lower than optimal.

### Where is this pain *really* coming from?

While the pain in my right hand had reduced considerably, it had not disppeared and would still occasionally flare up - especially with more mouse use. Curiously, I discovered that going back to the K860 actually helped reduce the pain. Something about the Kinesis and using the mouse a lot did not play nicely together. At a point the pain became sharp at the wrist, which is when I realized the likely cause. Have a look at the below image, which shows the K860 on the left, versus the outer edge of the Kinesis Advantage 360 on the right, with a mouse in the middle:

![height](/content/kinesis-mouse-k860.jpeg)

The height at which your hands rest is significantly higher when using the Kinesis versus any other keyboard. I have a standing desk that I had adjusted to a comfortable typing height *for the Kinesis*, which is roughly at 90 degree angle at the elbows, 0 at the wrists. At this table height, the mouse sits a good 4-5cm lower than it does with a flatter keyboard; when your hand reaches the mouse, your right elbow needs to extend, e.g. to 110 degress - however the surface of the desk is flat, so now your *wrist* has to compensate by flexing that 20 degrees upwards, leaving it in a strange angle.

I fix this by keeping a textbook as my mouse pad:

![height with textbook](/content/kinesis-with-textbook.jpeg)

## 6 months in

With the strange keyboard (and using a textbook mousepad), most of the pains I had previously have either disappeared or greatly reduced. I finally feel that I am no longer crippling my ability and speed to interact with a computer, and at times forget the existence of the keyboard entirely.

## Is it worth the effort?

With the knowledge I have today, I would prioritise everything about keyboard/mouse ergonomics roughly as follows in terms of importance:

1. **Reducing mouse usage is the most important thing**. For me, some mice hurt less than others, but all of them are painful longer-term. Get a wireless mouse you can turn off, and keep it turned off by default, and use the keyboard for everything.
2. **Slightly split/wavy layout** has a big effect on your comfort, with a minimal learning curve. [Logitech Ergo K860](https://amzn.to/4hqq3fK) is the best option I know.
3. **Fancier stuff like the columnar layouts and keywells** that the Kinesis Advantage 360 has can help further, but compared to the first two priorities, the effects are rather small and the learning curves are steep. It isn't worth the effort unless you've already addressed #1 and #2, and are looking to go beyond that.

{{< adsense >}}

