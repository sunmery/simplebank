import {
	Accordion,
	AccordionDetails,
	AccordionGroup,
	AccordionSummary,
	Link as RouterLink,
} from '@mui/joy'
import {useState} from 'react'

interface ProductItem {
	name: string
	url: string
}

interface Product {
	group: string
	item: ProductItem[]
}

/**
 *@returns JSXElement
 */
export default function Products() {
	const [products] = useState<Product[]>([
		{
			group: 'Cloud Native Storage',
			item: [
				{
					name: 'Longhorn',
					url: 'http://node6.api-r.com:32016/#/dashboard',
				},
				{
					name: 'Minio',
					url: 'https://node10.api-r.com:31377/',
				},
				{
					name: 'Postgres',
					url: '',
				},
				{
					name: 'Redis',
					url: '',
				},
			],
		},
		{
			group: 'Cloud Native Container Registry',
			item: [
				{
					name: 'TCR',
					url: 'https://console.cloud.tencent.com/tcr/repository',
				},
			],
		},
		{
			group: 'Cloud Native Logging',
			item: [
				{
					name: 'Loki',
					url: 'https://looke.grafana.net/a/grafana-lokiexplore-app/explore',
				},
			],
		},
		{
			group: 'Cloud Native Observability',
			item: [
				{
					name: 'Jaeger',
					url: 'http://node9.api-r.com:31135/',
				},
				{
					name: 'GitlabCI',
					url: 'https://gitlab.com/',
				},
				{
					name: 'Grafana',
					url: 'https://looke.grafana.net',
				},
				{
					name: 'Komodor',
					url: 'https://app.komodor.com/',
				},
			],
		},
		{
			group: 'Cloud Native Streaming & Messaging',
			item: [
				{
					name: 'Kafka',
					url: 'http://node5.api-r.com:31534/ui',
				},
			],
		},
		{
			group: 'Cloud Native Gateway',
			item: [
				{
					name: 'Higress',
					url: 'http://node8.api-r.com:32106/',
				},
			],
		},
		{
			group: 'Cloud Native Service Mesh',
			item: [
				{
					name: 'Consul',
					url: 'http://node6.api-r.com:31080/ui/dc1/services',
				},
			],
		},

		{
			group: 'Cloud Native Continuous Security & Compliance',
			item: [
				{
					name: 'Casdoor',
					url: 'http://node6.api-r.com:31550/',
				},
			],
		},
		{
			group: 'Cloud Native Continuous Integration & Delivery',
			item: [
				{
					name: 'ArgoCD',
					url: 'https://node6.api-r.com:32049/',
				},
				{
					name: 'GitlabCI',
					url: 'https://gitlab.com/',
				},
			],
		},
	])
	return (
		<AccordionGroup
			sx={{
				width: '30%',
			}}
			color="primary"
			size="lg"
			variant="plain"
		>
			{products.map((product: Product, index: number) => {
				return (
					<Accordion key={index}>
						<AccordionSummary>{product.group}</AccordionSummary>
						{product.item.map((item: ProductItem, index: number) => {
							return (
								<AccordionDetails key={index}>
									<RouterLink href={item.url}>{item.name}</RouterLink>
								</AccordionDetails>
							)
						})}
					</Accordion>
				)
			})}
		</AccordionGroup>
	)
}
