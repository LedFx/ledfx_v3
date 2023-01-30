=====================
 Welcome to LedFx V3
=====================
|Website| |License| |Discord| |Contributor Covenant| |Deepsource| |CodeCov|

.. image:: https://raw.githubusercontent.com/LedFx/LedFx/b8e68beaa215d4308c74d0c7d657556ac894b707/icons/banner.png

What is LedFx? Music to Light!
------------------------------

LedFx makes a live light show from your music, creating a more immersive experience for you.

The visual effects are primarily written for LED strips, but can also work with bulbs and LED panels.

A cut above the rest
--------------------

Off-the-shelf music reactive LED strips have poor quality microphones and lazy music processing algorithms. The result is jerky and uncomfortable lighting that doesn't match the atmosphere of your music.

The other option is to pre-program your lights for your music, like a concert. This requires a lot of effort to get good looking results.

LedFx solves both of these issues, performing tonal and rhythmic analysis on your music before it even leaves your speakers. The result is an effortlessly stunning light show for you to enjoy. 

Smart Connectivity
------------------

.. image:: https://user-images.githubusercontent.com/32398028/215608580-29743af4-cf72-409c-9261-4eb51d92c659.png

LedFx appears as a smart speaker on your devices. Add it to your speaker groups to feed it your music.

Match the Musical Mood
----------------------

In conjunction with realtime audio analysis, LedFx uses Spotify's API to determine the mood and segmentation of your music. It knows when there's a beat drop coming up, and will show this in its lighting effects.

.. code-block:: json

  {
    "acousticness": 0.00242,
    "danceability": 0.585,
    "energy": 0.842,
    "instrumentalness": 0.00686,
    "key": 9,
    "liveness": 0.0866,
    "loudness": -5.883,
    "mode": 0,
    "speechiness": 0.0556,
    "tempo": 118.211,
    "time_signature": 4,
    "valence": 0.428
  }
  
Installation
------------

At present, this project is in pre-release development, and there are no packages available for download.

Check out `LedFx V2 <https://github.com/ledfx/ledfx>`_ if you want to get your lights dancing right away!

If you are a developer and want to contribute, get in touch with the development team via Discord and we'll help you get started.

.. |Discord| image:: https://img.shields.io/badge/chat-on%20discord-7289da.svg
   :target: https://discord.gg/xyyHEquZKQ
   :alt: Discord
.. |Contributor Covenant| image:: https://img.shields.io/badge/Contributor%20Covenant-v2.0%20adopted-ff69b4.svg
   :target: CODE_OF_CONDUCT.md
.. |License| image:: https://img.shields.io/badge/License-AGPLv3-blue
   :alt: License
.. |Website| image:: https://img.shields.io/website?down_color=red&down_message=is%20unavailable&up_color=green&up_message=more%20info&url=https%3A%2F%2Fledfx.app
   :alt: Our Website
.. |Deepsource| image:: https://deepsource.io/gh/LedFx/ledfx_v3.svg/?label=active+issues&show_trend=true&token=E2DuDD9meHHrq-jZVKtzHW4a
  :target: https://deepsource.io/gh/LedFx/ledfx_v3/?ref=repository-badge
.. |CodeCov| image:: https://codecov.io/gh/LedFx/ledfx_v3/branch/main/graph/badge.svg
:target: https://codecov.io/gh/LedFx/ledfx_v3


