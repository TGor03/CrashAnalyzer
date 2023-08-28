let serverURL = "localhost:25478" // <- Change this to your server URL or IP format is (IP:PORT)

// ************************ Drag and drop ***************** //
let dropArea = document.getElementById("filebox")
let fileplus = document.getElementById("fileplus")

// Prevent default drag behaviors
;['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
  dropArea.addEventListener(eventName, preventDefaults, false)   
  document.body.addEventListener(eventName, preventDefaults, false)
})

// Highlight drop area when item is dragged over it
;['dragenter', 'dragover'].forEach(eventName => {
  dropArea.addEventListener(eventName, highlight, false)
})

;['dragleave', 'drop'].forEach(eventName => {
  dropArea.addEventListener(eventName, unhighlight, false)
})

// Handle dropped files
dropArea.addEventListener('drop', handleDrop, false)

function preventDefaults (e) {
  e.preventDefault()
  e.stopPropagation()
}

function highlight(e) {
  dropArea.classList.add('highlight')
  fileplus.classList.add('highlight')
}

function unhighlight(e) {
  dropArea.classList.remove('highlight')
  fileplus.classList.remove('highlight')
}

function handleDrop(e) {
  var dt = e.dataTransfer
  var files = dt.files

  handleFiles(files)
}

let closebutton = document.getElementById('closebutton')
let outputbox = document.getElementById('output')
let spinner = document.getElementById('loading')


function handleFiles(files) {
  files = [...files]
  files.forEach(uploadFile)
}

// Point at which the file is uploaded
function uploadFile(file, i) {
  var url = 'http://' + serverURL + '/files/dump.dmp'
  var xhr = new XMLHttpRequest()
  var formData = new FormData()
  xhr.open('PUT', url, true)
  xhr.setRequestHeader('X-Requested-With', 'XMLHttpRequest')

  // Update progress (can be used to show progress indicator)
  xhr.upload.addEventListener("progress", function(e) {
    dropArea.classList.add('hidden')
    fileplus.classList.add('hidden')
    spinner.classList.add('shown')
    closebutton.classList.add('shown')
  })

  xhr.addEventListener('readystatechange', function(e) {
    if (xhr.readyState == 4 && xhr.status == 200) {
      dropArea.classList.add('hidden')
      fileplus.classList.add('hidden')
      spinner.classList.remove('shown')
      outputbox.innerText = xhr.responseText
    }
    else if (xhr.readyState == 4 && xhr.status != 200) {
      alert('Error uploading file please check console!')
    }
  })

  formData.append('file', file)
  xhr.send(formData)
  xhr.getAllResponseHeaders
}

// Close button
function closeDump() {
  //Just reload the page
  location.reload();
}