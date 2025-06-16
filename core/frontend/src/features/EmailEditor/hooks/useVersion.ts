export const useVersion = (data: string) => {
	const transformData = (data: string) => {
		try {
			const parsedData = JSON.parse(data)
			return parsedData
		} catch {
			// console.log(data, err)
			return null
		}
	}

	return {
		parsedData: transformData(data),
	}
}
