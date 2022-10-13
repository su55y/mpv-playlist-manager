'use strict'
const TITLE = 'Add to MPV',
  TUTORIAL = 'https://youtu.be/dQw4w9WgXcQ',
  APPEND_URL = 'http://localhost:5000/req?u=',
  YOUTUBE_ICON = 'icons/yt48.png',
  TWITCH_ICON = 'icons/tw48.png',
  MPV_ICON = 'icons/mpv48.png',
  REGEXES = [
    {
      icon: YOUTUBE_ICON,
      rx: /^.*youtube\.com\/watch\?v=([a-zA-Z0-9\-_]{11})\&?.*$/,
      len: 11,
    },
    {
      icon: YOUTUBE_ICON,
      rx: /^.*youtube\.com\/shorts\/([a-zA-Z0-9\-_]{11})$/,
      len: 11,
    },
    { icon: TWITCH_ICON, rx: /^.*twitch\.tv\/([^\?\/]+)$/ },
    {
      icon: TWITCH_ICON,
      rx: /^.*twitch\.tv\/videos\/([\d]{10})$/,
      len: 10,
    },
    { icon: TWITCH_ICON, rx: /^(.*twitch.tv\/[^\/]+\/clip\/.+)$/ },
  ]

const notifySend = (message, icon) => {
  chrome.notifications.create({
    type: 'basic',
    iconUrl: icon,
    title: TITLE,
    message,
    priority: 0,
  })
}

const appendVideo = async (u) => {
  const response = await fetch(`${APPEND_URL}${encodeURIComponent(u)}`)
  const result = await response.json()
  return result?.error === 'success'
}

const parseUrl = (url) => {
  let msg
  for (const { rx, icon, len } of REGEXES) {
    if (!url.match(rx)) continue
    const res = url.match(rx)
    if (len && res[1].length !== len) {
      msg = `invalid url(${url})`
      console.log(msg)
      return { url, msg, icon, ok: false }
    }
    msg = `${url} video just added`
    console.log(msg)
    return { url: res[0], msg, icon, ok: true }
  }
  msg = 'unknown url'
  return { url, msg, icon: MPV_ICON, ok: false }
}

const onClickContextMenu = (info) => {
  const { pageUrl, linkUrl } = info
  const { url, msg, icon, ok } = parseUrl(linkUrl || pageUrl)
  if (!url || !ok) {
    notifySend(msg, icon)
    return
  }
  console.log(`send ${url}`)
  if (appendVideo(url)) notifySend(msg, icon)
}

;['page', 'link'].forEach((e, i) => {
  chrome.contextMenus.create({
    id: `e${i}`,
    title: TITLE,
    contexts: [e],
  })
})

chrome.contextMenus.onClicked.addListener(onClickContextMenu)
