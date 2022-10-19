'use strict'
const PORT = '5000',
  HEALTHZ_URL = `http://localhost:${PORT}/healthz`,
  PLAYLIST_URL = `http://localhost:${PORT}/playlist`,
  CONTROL_URL = `http://localhost:${PORT}/control?action=`,
  PLAY_INDEX_URL = `http://localhost:${PORT}/play?index=`,
  controlButtons = document.getElementsByClassName('control_button'),
  updButton = document.getElementById('upd_btn'),
  OK = 'success',
  CROSS_SVG = `<svg viewBox="0 0 10 10" width="15" height="15">
  <polygon points="0,1 1,0 10,9 9,10" />
  <polygon points="0,9 1,10 10,1 9,0" />
</svg>`

const refreshList = (playlist) => {
  if (playlist) {
    let playlistElement = document.createElement('div')
    let btns = []
    for (const [i, v] of playlist.entries()) {
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
      if (!v.current) {
        const id = `play${i}`
        title.id = id
        btns.push(id)
        title.setAttribute('data-index', i)
      }
      el.appendChild(title)
      let buttonsCont = document.createElement('div')
      buttonsCont.className = 'pl_item_buttons'
      let removeBtn = document.createElement('button')
      removeBtn.id = v.id
      removeBtn.className = 'remove_button'
      removeBtn.innerHTML = CROSS_SVG
      buttonsCont.appendChild(removeBtn)
      el.appendChild(buttonsCont)

      playlistElement.appendChild(el)
    }
    document.getElementById('playlist').innerHTML = playlistElement.innerHTML
    for (const b of btns)
      document.getElementById(b).addEventListener('click', playIndex)
  }
  updButton.disabled = false
}

const playIndex = async (e) => {
  e.target.disabled = true
  const id = e.target.dataset?.index || false
  if (!id) return
  await fetch(`${PLAY_INDEX_URL}${id}`)
    .then((r) => r.json())
    .then((r) => {
      if (r.error === OK) refreshPlaylist()
    })
}

const refreshPlaylist = async () => {
  updButton.disabled = true
  await fetch(PLAYLIST_URL)
    .then((r) => r.json())
    .then(refreshList)
}

const control = async (id) => {
  for (const b of controlButtons) b.disabled = true
  await fetch(`${CONTROL_URL}${id}`)
    .then((r) => r.json())
    .then((r) => {
      if (r.error === OK) refreshPlaylist()
      for (const b of controlButtons) b.disabled = false
    })
}

updButton.addEventListener('click', refreshPlaylist)
for (const e of controlButtons) {
  e.addEventListener('click', () => control(e.id))
}

window.onload = async function () {
  await fetch(HEALTHZ_URL).then((r) => {
    if (r.status === 204) refreshPlaylist()
    else alert('servers is down')
  })
}
