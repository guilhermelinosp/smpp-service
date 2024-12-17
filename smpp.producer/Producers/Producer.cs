using System.Text;
using RabbitMQ.Client;

namespace smpp.producer.Producers;

public class Producer(IConfiguration configuration)
{
	private static IChannel? _channel;
	private static string? _queueName;

	public async Task InitializeAsync(CancellationToken cancellationToken)
	{
		cancellationToken.ThrowIfCancellationRequested();

		_queueName = configuration["RabbitMQ:QueueName"] ??
		             throw new InvalidOperationException("Queue name cannot be null or empty.");

		var connectionFactory = new ConnectionFactory
		{
			HostName = configuration["RabbitMQ:HostName"] ??
			           throw new InvalidOperationException("HostName cannot be null."),
			Port = int.TryParse(configuration["RabbitMQ:Port"], out var port)
				? port
				: throw new InvalidOperationException("Invalid Port."),
			UserName = configuration["RabbitMQ:UserName"] ??
			           throw new InvalidOperationException("UserName cannot be null."),
			Password = configuration["RabbitMQ:Password"] ??
			           throw new InvalidOperationException("Password cannot be null."),
		};

		var connection = await connectionFactory.CreateConnectionAsync(cancellationToken);

		_channel = await connection.CreateChannelAsync(null, cancellationToken);

		await _channel.QueueDeclareAsync(
			queue: _queueName,
			durable: false,
			exclusive: false,
			autoDelete: false,
			arguments: null,
			cancellationToken: cancellationToken);
	}

	public static async Task SendAsync(string message, CancellationToken cancellationToken)
	{
		cancellationToken.ThrowIfCancellationRequested();

		var body = Encoding.UTF8.GetBytes(message);
		await _channel!.BasicPublishAsync(
			exchange: string.Empty,
			routingKey: _queueName!,
			body: body,
			cancellationToken: cancellationToken);
	}
}