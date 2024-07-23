const fileQueryResponsePromises = {}

const testExistingFileQuery = async () => {
  // Query this file since we know it should exists
  // We can only quey in the scope of our current plugin
  const fileQuery = {
    procedure: "FILE_QUERY",
    args: {
      path: "processors/wasm_worker_protocol.js",
    },
  }

  // Build the promise ready that will return the file
  let resolve, reject, promise = new Promise((res, rej) => {
    resolve = res
    reject = rej
  })
  fileQueryResponsePromises[fileQuery.args.path] = {
    promise,
    resolve,
    reject
  }

  // Send the file query message
  self.postMessage(fileQuery)

  // Wait for the response and check the result
  const fileResponse = await promise
  if(!fileResponse instanceof Blob) {
    self.postMessage({
      procedure: "ERROR",
      args: {
        exception: `Invalid response message received on existing file query: ${fileResponse}`,
      },
    })
    return false
  }

  return true
}

const testNonExistingFileQuery = async () => {
  // Query this file since we know it should exists
  // We can only quey in the scope of our current plugin
  const fileQuery = {
    procedure: "FILE_QUERY",
    args: {
      path: "processors/non_existing_file.txt",
    },
  }

  // Build the promise ready that will return the file
  let resolve, reject, promise = new Promise((res, rej) => {
    resolve = res
    reject = rej
  })
  fileQueryResponsePromises[fileQuery.args.path] = {
    promise,
    resolve,
    reject
  }

  // Send the file query message
  self.postMessage(fileQuery)

  // Wait for the response and check the result
  const fileResponse = await promise
  if(fileResponse != undefined) {
    self.postMessage({
      procedure: "ERROR",
      args: {
        exception: `Invalid response message received on non existing file query: ${fileResponse}`,
      },
    })
    return false
  }

  return true
}

// Handler for message to this web worker
self.onmessage = async (event) => {
  try {
    const {procedure, args} = event.data
    switch (procedure) {
      // Send some testing FILE_QUERY requests to see if we receive the expected files
      case "INIT":
        if(!await testExistingFileQuery()) {
          break
        }
        if(!await testNonExistingFileQuery()) {
          break
        }

        break

      case "FILE_RESPONSE":
        // Resolve the pending promise
        const {path, content} = args || {}
        if(path in fileQueryResponsePromises) {
          fileQueryResponsePromises[path].resolve(content)
        } else {
          self.postMessage({
            procedure: "ERROR",
            args: {
              exception: `File response at path ${JSON.stringify(event.data)} was not expected`
            }
          })
        }
        break;

      case "SDP_ANSWER":
        

      default:
        break;
    }
  } catch (e) {
    self.postMessage({
      procedure: "ERROR",
      args: {
        exception: e.stack
      }
    })
  }
}

