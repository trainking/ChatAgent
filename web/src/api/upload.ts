import axios from 'axios'

const token = () => localStorage.getItem('token') || ''

export function uploadFile(file: File) {
  const formData = new FormData()
  formData.append('file', file)
  return axios.post('/api/v1/upload', formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
      Authorization: `Bearer ${token()}`,
    },
  })
}
