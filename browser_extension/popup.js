'use strict'
const HEALTHZ_URL = 'http://localhost:5000/healthz',
  PLAYLIST_URL = 'http://localhost:5000/playlist',
  CONTROL_URL = 'http://localhost:5000/control?action=',
  controlButtons = document.getElementsByClassName('control_button'),
  updButton = document.getElementById('upd_btn')

const refreshList = (playlist) => {
  if (playlist) {
    let playlistElement = document.createElement('div')
    for (const v of playlist) {
      console.log(v)
      let el = document.createElement('div')
      el.className = 'pl_item'
      el.id = `pl_item${v.id}`
      let title = document.createElement('div')
      title.className = 'pl_item_title'
      if (v.current) el.className += ' pl_item_current'
      if (v.title) title.innerHTML = v.title
      else title.innerHTML = v.filename
      if (v.thumbnail) {
        let thumbCont = document.createElement('div'),
          thumb = new Image(100, 60)
        thumbCont.className = 'pl_item_thumb'
        thumb.src = v.thumbnail
        thumbCont.appendChild(thumb)
        el.appendChild(thumbCont)
      }
      let buttonsCont = document.createElement('div')
      buttonsCont.className = 'pl_item_buttons'
      let removeBtn = document.createElement('button')
      removeBtn.id = v.id
      removeBtn.className = 'remove_button'
      removeBtn.innerHTML = '\uf00d'
      if (!v.playing) {
        let playBtn = document.createElement('button')
        playBtn.id = v.id
        playBtn.className = 'play_button'
        playBtn.innerHTML = '\uf04b'
        buttonsCont.appendChild(playBtn)
      }
      buttonsCont.appendChild(removeBtn)
      el.appendChild(title)
      el.appendChild(buttonsCont)
      playlistElement.appendChild(el)
    }
    document.getElementById('playlist').innerHTML = playlistElement.innerHTML
  }
  updButton.disabled = false
}
const refreshPlaylist = async () => {
  updButton.disabled = true
  await fetch(PLAYLIST_URL)
    .then((r) => r.json())
    .then(refreshList)
}

const control = async (e) => {
  console.log(`action: ${e.target.value}`)
  for (const b of controlButtons) b.disabled = true
  await fetch(`${CONTROL_URL}${e.target.value}`)
    .then((r) => r.json())
    .then((r) => {
      if (r.error === 'success') console.log('success')
      for (const b of controlButtons) b.disabled = false
    })
}

updButton.addEventListener('click', refreshPlaylist)
for (const e of controlButtons) {
  e.addEventListener('click', control)
}

window.onload = async function () {
  await fetch(HEALTHZ_URL).then((r) => {
    if (r.status === 204) refreshPlaylist()
    else alert('servers is down')
  })
}
