'use strict'
document
  .getElementById('hi_btn')
  .addEventListener('click', (e) => alert(`btn.value = ${e.target.value}`))
