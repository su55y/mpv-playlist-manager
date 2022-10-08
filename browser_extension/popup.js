'use strict'
const PLAYLIST_URL = 'http://localhost:5000/playlist'

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
      el.appendChild(title)
      playlistElement.appendChild(el)
    }
    document.getElementById('playlist').innerHTML = playlistElement.innerHTML
  }
}
const refreshPlaylist = async () => {
  await fetch(PLAYLIST_URL)
    .then((r) => r.json())
    .then(refreshList)
}

document.getElementById('upd_btn').addEventListener('click', refreshPlaylist)
window.onload = function () {
  refreshPlaylist()
}
