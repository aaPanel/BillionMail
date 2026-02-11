import { instance } from '@/api'

export const getScraperStatus = () => {
	return instance.get('/frostbyte/proxy/leads/pipeline')
}

export const startScrapeJob = (data: { query: string; location: string; max_results: number }) => {
	return instance.post('/frostbyte/proxy/leads/scrape', data)
}
