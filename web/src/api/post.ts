import request from './index'

export interface PostItem {
  id: number
  code: string
  name: string
  sort: number
  status: number
  createdAt: string
}

export interface PostQuery {
  page: number
  pageSize: number
}

export function getPostList(params: PostQuery) {
  return request.get('/system/post/list', { params })
}

export function createPost(data: Partial<PostItem>) {
  return request.post('/system/post', data)
}

export function updatePost(data: Partial<PostItem>) {
  return request.put('/system/post', data)
}

export function deletePost(id: number) {
  return request.delete(`/system/post/${id}`)
}
