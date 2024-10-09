import type {ChangeEvent} from 'react'
import {useState} from 'react'
import {skipToken, useQuery} from '@tanstack/react-query'
import {Alert} from '@mui/joy'

interface RegisterUser {
	username: string
	fullName: string
	email: string
	password: string
}

const createUser = async (user: RegisterUser) => {
	try {
		const res = await fetch('http://localhost:8080/users', {
			method: 'PUT',
			headers: {'Content-Type': 'application/json'},
			body: JSON.stringify({...user}),
		})
		const data: RegisterUser = await res.json()
		return data
	} catch (error) {
		if (error instanceof Error) {
			throw new Error(error.message)
		}
		throw new Error(String(error))
	}
}
interface User {
	username: string
	fullName: string
	email: string
	password: string
}

/**
 * @returns JSXElement
 */
export default function Register() {
	const [user, setUser] = useState<User>({
		username: '',
		fullName: '',
		email: '',
		password: '',
	})
	const [query, setQuery] = useState<boolean>(false)

	const {isError, data, error} = useQuery({
		queryKey: ['register', user],
		queryFn: query ? () => createUser(user) : skipToken,
	})

	if (isError) {
		return <span>Error: {error.message}</span>
	}

	const handleRegister = () => {
		setQuery(true)
	}

	const handleInputChange = (e: ChangeEvent<HTMLInputElement>) => {
		const {name, value} = e.target
		setUser((prevState) => ({
			...prevState,
			[name]: value,
		}))
	}

	if (data) {
		return (
			<Alert
				variant="solid"
				color="success"
			>
				注册成功! 欢迎您 {data.fullName}
			</Alert>
		)
	}

	return (
		<>
			<label htmlFor="">
				<input
					type="text"
					name="username"
					id=""
					value={user.username}
					onChange={handleInputChange}
				/>
			</label>
			<label htmlFor="">
				<input
					type="text"
					name="fullName"
					id=""
					onChange={handleInputChange}
				/>
			</label>
			<label htmlFor="">
				<input
					type="email"
					name="email"
					id=""
					onChange={handleInputChange}
				/>
			</label>

			<label htmlFor="">
				<input
					type="password"
					name="password"
					id=""
					onChange={handleInputChange}
				/>
			</label>
			<label htmlFor="">
				重复密码:
				<input
					type="password"
					name="repPassword"
					id=""
				/>
			</label>
			<button
				type="button"
				onClick={handleRegister}
			>
				Register
			</button>
		</>
	)
}
