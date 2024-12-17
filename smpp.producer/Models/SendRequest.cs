using Newtonsoft.Json;

namespace smpp.producer.Models;

public record SendRequest
{
	[JsonProperty("to")] 
	public string? To { get; set; }
	
	[JsonProperty("message")] 
	public string? Message { get; set; }
	
	[JsonProperty("priority")] 
	public string? Priority { get; set; }
}