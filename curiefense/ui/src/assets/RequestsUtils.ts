import axios, {AxiosRequestConfig, AxiosResponse} from 'axios'
import DatasetsUtils from '@/assets/DatasetsUtils'
import * as bulmaToast from 'bulma-toast'
import {ToastType} from 'bulma-toast'

export type MethodNames = 'GET' | 'PUT' | 'POST' | 'DELETE'

const apiRoot = DatasetsUtils.ConfAPIRoot
const apiVersion = DatasetsUtils.ConfAPIVersion
const logsApiRoot = DatasetsUtils.LogsAPIRoot
const logsApiVersion = DatasetsUtils.LogsAPIVersion
const axiosMethodsMap: Record<MethodNames, Function> = {
  'GET': axios.get,
  'PUT': axios.put,
  'POST': axios.post,
  'DELETE': axios.delete,
}

const toast = (message: string, type: ToastType) => {
  bulmaToast.toast({
    message: message,
    type: <ToastType>`is-light ${type}`,
    position: 'bottom-left',
    closeOnClick: true,
    pauseOnHover: true,
    duration: 3000,
    opacity: 0.8,
  })
}

const successToast = (message: string) => {
  toast(message, 'is-success')
}

const failureToast = (message: string) => {
  toast(message, 'is-danger')
}

const processRequest = (methodName: MethodNames, apiUrl: string, data: any, config: AxiosRequestConfig,
                        successMessage: string, failureMessage: string) => {
  // Get correct axios method
  if (!methodName) {
    methodName = 'GET'
  } else {
    methodName = <MethodNames>methodName.toUpperCase()
  }
  const axiosMethod = axiosMethodsMap[methodName]
  if (!axiosMethod) {
    console.error(`Attempted sending unrecognized request method ${methodName}`)
    return
  }

  // Request
  console.log(`Sending ${methodName} request to url ${apiUrl}`)
  let request
  if (data) {
    if (config) {
      request = axiosMethod(apiUrl, data, config)
    } else {
      request = axiosMethod(apiUrl, data)
    }
  } else {
    request = axiosMethod(apiUrl)
  }
  request = request.then((response: AxiosResponse) => {
    // Toast message
    if (successMessage) {
      successToast(successMessage)
    }
    return response
  }).catch((error: Error) => {
    // Toast message
    if (failureMessage) {
      failureToast(failureMessage)
    }
    throw error
  })
  return request
}

const sendRequest = (methodName: MethodNames, urlTail: string, data?: any, config?: AxiosRequestConfig,
                     successMessage?: string, failureMessage?: string) => {
  const apiUrl = `${apiRoot}/${apiVersion}/${urlTail}`
  return processRequest(methodName, apiUrl, data, config, successMessage, failureMessage)
}

const sendLogsRequest = (methodName: MethodNames, urlTail: string, data?: any, config?: AxiosRequestConfig,
                         successMessage?: string, failureMessage?: string) => {
  const apiUrl = `${logsApiRoot}/${logsApiVersion}/${urlTail}`
  return processRequest(methodName, apiUrl, data, config, successMessage, failureMessage)
}

export default {
  name: 'RequestsUtils',
  sendRequest,
  sendLogsRequest,
}
